package xnet

type Handler interface {
	Connect(conn Conn)
	Packet(conn Conn, pkt Packet)
	Disconnect(conn Conn)
}
