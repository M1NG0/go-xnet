package xnet

type Packet interface {
	Marshal() []byte
	Unmarshal([]byte) error
}

type Protocal interface {
	Pack(pkt Packet) (b []byte)
	Unpack(b []byte) (pkt Packet, len int, err error)
}
