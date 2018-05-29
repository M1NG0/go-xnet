package xnet

import (
	"net"
)

type Conn interface {
	LocalAddr() (laddr net.Addr)
	RemoteAddr() (raddr net.Addr)

	Write(pkt Packet) (err error)
	AsyncWrite(pkt Packet) (err error)
	Close() (err error)
}
