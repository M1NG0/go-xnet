package xnet

import (
	"context"
	"net"
	"sync"
	"time"
)

const MaxDatagramSize = 64 * 1024

type UDPServer struct {
	Network         string
	Addr            string
	Protocal        Protocal
	Handler         Handler
	ConnectTimeout  time.Duration
	MaxDatagramSize int

	listener *net.UDPConn
	wg       *sync.WaitGroup
	closed   chan struct{}

	conns *sync.Map
}

const DefaultConnectTimeout = 10 * time.Minute

func NewUDPServer(network, addr string, protocal Protocal, handler Handler) (srv *UDPServer) {
	return &UDPServer{
		Addr:            addr,
		Network:         network,
		Protocal:        protocal,
		Handler:         handler,
		ConnectTimeout:  DefaultConnectTimeout,
		MaxDatagramSize: MaxDatagramSize,
	}
}

func (srv *UDPServer) Serve() (err error) {
	laddr, err := net.ResolveUDPAddr(srv.Network, srv.Addr)
	if err != nil {
		return err
	}
	listener, err := net.ListenUDP(srv.Network, laddr)
	if err != nil {
		return err
	}
	srv.listener = listener
	srv.wg = &sync.WaitGroup{}
	srv.conns = &sync.Map{}
	srv.closed = make(chan struct{})
	return srv.loop()
}

func (srv *UDPServer) Shutdown(ctx context.Context) (err error) {
	close(srv.closed)
	finish := make(chan struct{})
	go func() {
		srv.wg.Wait()
		close(finish)
	}()
	select {
	case <-ctx.Done():
	case <-finish:
	}
	return srv.listener.Close()
}

func (srv *UDPServer) Close() (err error) {
	close(srv.closed)
	return srv.listener.Close()
}

func (srv *UDPServer) loop() (err error) {
	// srv.wg.Add(1)
	// defer srv.wg.Done()
	buf := make([]byte, srv.MaxDatagramSize)
	for {
		n, addr, err := srv.listener.ReadFrom(buf)
		if err != nil {
			// if ok {
			// 	conn.close()
			// }
			break
		}
		val, ok := srv.conns.LoadOrStore(addr.Network(), newUDPConn(srv, addr))
		conn := val.(*UDPConn)

		if !ok {
			go conn.loop()
		}
		datagram := make([]byte, n)
		copy(datagram, buf[:n])
		select {
		case <-conn.closed:
			goto DONE
		case conn.datagramChan <- datagram:
		}
	}
DONE:
	return nil
}
