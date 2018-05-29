package xnet

import (
	"context"
	"fmt"
	"testing"
	"time"
)

type testPacket struct {
	Body string
}

func (pkt *testPacket) Marshal() []byte {
	return []byte(pkt.Body)
}
func (pkt *testPacket) Unmarshal(b []byte) error {
	pkt.Body = string(b)
	return nil
}

type testProtocal struct {
}

func (p *testProtocal) Pack(pkt Packet) []byte {
	return pkt.Marshal()
}

func (p *testProtocal) Unpack(b []byte) (Packet, int, error) {
	fmt.Println("receive data: ", string(b))
	pkt := new(testPacket)
	pkt.Unmarshal(b)
	return pkt, len(b), nil
}

type testHandler struct {
}

func (h *testHandler) Connect(conn Conn) {
	fmt.Println("connect:", conn.RemoteAddr(), conn.LocalAddr())
}

func (h *testHandler) Packet(conn Conn, pkt Packet) {
	fmt.Println("packet:", string(pkt.Marshal()))
	err := conn.Write(&testPacket{Body: "hello"})
	if err != nil {
		panic(err)
	}
}

func (h *testHandler) Disconnect(conn Conn) {
	fmt.Println("disconnect:", conn.RemoteAddr())
}

func TestUDPServer(t *testing.T) {
	srv := NewUDPServer("udp", "192.168.1.7:5060", &testProtocal{}, &testHandler{})
	go func() {
		err := srv.Serve()
		if err != nil {
			t.Error(err)
			t.Fail()
		}
	}()
	time.Sleep(1 * time.Minute)
	err := srv.Shutdown(context.Background())
	if err != nil {
		t.Error(err)
		t.Fail()
	}
}
