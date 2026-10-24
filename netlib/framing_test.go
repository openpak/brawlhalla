package netlib

import (
	"bytes"
	"testing"
)

func TestFramingRoundTrip(t *testing.T) {
	b := New(ClientServerTcpHandshake)
	b.W.String("Brawlhalla client to server protocol 1.0")
	p := b.Finish()
	raw, err := p.Encode()
	if err != nil {
		t.Fatal(err)
	}
	// type 178, length 42 (u16 string length + 40 bytes), big-endian.
	if !bytes.Equal(raw[:4], []byte{0x00, 0xb2, 0x00, 0x2a}) {
		t.Fatalf("header = % x", raw[:4])
	}
	login := &Packet{Type: 10400, Sender: 0x01020304, Payload: []byte{9}}
	raw2, _ := login.Encode()
	if !bytes.Equal(raw2, []byte{0xa8, 0xa0, 1, 2, 3, 4, 0, 1, 9}) {
		t.Fatalf("sender header = % x", raw2)
	}
	r := bytes.NewReader(append(raw, raw2...))
	got, err := ReadPacket(r)
	if err != nil || got.Type != ClientServerTcpHandshake || NewReader(got.Payload).String() != "Brawlhalla client to server protocol 1.0" {
		t.Fatalf("first packet: %+v %v", got, err)
	}
	got, err = ReadPacket(r)
	if err != nil || got.Type != 10400 || got.Sender != 0x01020304 || !bytes.Equal(got.Payload, []byte{9}) {
		t.Fatalf("second packet: %+v %v", got, err)
	}
	if _, err := ReadPacket(bytes.NewReader([]byte{0, 1, 0})); err == nil {
		t.Fatal("a truncated header read")
	}
}
