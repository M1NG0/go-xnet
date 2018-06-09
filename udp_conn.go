package xnet

import (
	"net"
	"sync"
	"time"
)

type UDPConn struct {
	srv   *UDPServer
	raddr net.Addr

	datagramChan chan []byte
	// bufMux       *sync.Mutex

	wPktChan chan Packet
	rPktChan chan Packet
	closed   chan struct{}
}

func newUDPConn(srv *UDPServer, raddr net.Addr) *UDPConn {
	conn := &UDPConn{
		srv:   srv,
		raddr: raddr,

		datagramChan: make(chan []byte, 10),
		// bufMux:       &sync.Mutex{},

		wPktChan: make(chan Packet, 10),
		rPktChan: make(chan Packet, 10),
		closed:   make(chan struct{}),
	}
	return conn
}

func (conn *UDPConn) RemoteAddr() net.Addr {
	return conn.raddr
}

func (conn *UDPConn) LocalAddr() net.Addr {
	return conn.srv.listener.LocalAddr()
}

func (conn *UDPConn) Write(pkt Packet) (err error) {
	b := conn.srv.Protocal.Pack(pkt)
	_, err = conn.srv.listener.WriteTo(b, conn.raddr)
	return err
}

func (conn *UDPConn) AsyncWrite(pkt Packet) (err error) {
	conn.wPktChan <- pkt
	return
}

func (conn *UDPConn) Close() (err error) {
	return conn.close()
}

func (conn *UDPConn) close() (err error) {
	close(conn.closed)
	conn.srv.conns.Delete(conn.raddr.Network())
	conn.srv.Handler.Disconnect(conn)
	return nil
}

func (conn *UDPConn) loop() {
	conn.srv.Handler.Connect(conn)
	asyncDo(conn.writeLoop, conn.srv.wg)
	asyncDo(conn.readLoop, conn.srv.wg)
	asyncDo(conn.handleLoop, conn.srv.wg)
	select {
	case <-conn.closed:
	case <-conn.srv.closed:
	}
	conn.srv.Handler.Disconnect(conn)
}

func (conn *UDPConn) writeLoop() {
	for {
		select {
		case <-conn.closed:
			return
		case <-conn.srv.closed:
			return
		case pkt := <-conn.wPktChan:
			conn.Write(pkt)
		}
	}
}

func (conn *UDPConn) readLoop() {
	for {
		select {
		case <-conn.closed:
			return
		case <-conn.srv.closed:
			return
		case datagram := <-conn.datagramChan:
			pkt, _, err := conn.srv.Protocal.Unpack(datagram)
			if err == nil && pkt != nil {
				conn.rPktChan <- pkt
			}
		case <-time.After(conn.srv.ConnectTimeout):
			conn.Close()
		}
	}
}

func (conn *UDPConn) handleLoop() {
	for {
		select {
		case <-conn.closed:
			return
		case <-conn.srv.closed:
			return
		case pkt := <-conn.rPktChan:
			conn.srv.Handler.Packet(conn, pkt)
		}
	}
}

func asyncDo(fn func(), wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		fn()
		wg.Done()
	}()
}
