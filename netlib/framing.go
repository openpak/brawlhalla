package netlib

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

// Packet is one Blue Mammoth packet: a type and a bitstream payload. A few
// types also carry the sender's user id (HasSender).
type Packet struct {
	Type    uint16
	Sender  uint32
	Payload []byte
}

// MaxPayload is the largest payload a 16-bit length can carry.
const MaxPayload = 0xffff

// senderFlag marks a header that carries a 32-bit sender after the type. The
// client sets it for the three types Connection::SendPacket singles out:
// GAMESERVER_LOGIN (10400), GAMESERVER_LOGIN_REJOIN (10405) and
// CONNECT_SPECTATE (10202).
const senderFlag = 0x8000

// HasSender reports whether the client sends packets of type t with a sender.
func HasSender(t uint16) bool { return t == 10400 || t == 10405 || t == 10202 }

// The TCP stream is a run of packets, each
//
//	u16 type (bit 15 set: a u32 sender follows)
//	[u32 sender]
//	u16 payload length
//	payload
//
// with every integer big-endian (Connection::SendPacket).

// ReadPacket reads one packet from a stream.
func ReadPacket(r io.Reader) (*Packet, error) {
	var h [2]byte
	if _, err := io.ReadFull(r, h[:]); err != nil {
		return nil, err
	}
	p := &Packet{Type: binary.BigEndian.Uint16(h[:])}
	if p.Type&senderFlag != 0 {
		p.Type &^= senderFlag
		var s [4]byte
		if _, err := io.ReadFull(r, s[:]); err != nil {
			return nil, unexpected(err)
		}
		p.Sender = binary.BigEndian.Uint32(s[:])
	}
	if _, err := io.ReadFull(r, h[:]); err != nil {
		return nil, unexpected(err)
	}
	p.Payload = make([]byte, binary.BigEndian.Uint16(h[:]))
	if _, err := io.ReadFull(r, p.Payload); err != nil {
		return nil, unexpected(err)
	}
	return p, nil
}

func unexpected(err error) error {
	if errors.Is(err, io.EOF) {
		return io.ErrUnexpectedEOF
	}
	return err
}

// Encode is the packet as the stream carries it.
func (p *Packet) Encode() ([]byte, error) {
	if len(p.Payload) > MaxPayload {
		return nil, fmt.Errorf("netlib: packet %d payload is %d bytes", p.Type, len(p.Payload))
	}
	out := make([]byte, 0, 8+len(p.Payload))
	t := p.Type
	if p.Sender != 0 {
		t |= senderFlag
	}
	out = binary.BigEndian.AppendUint16(out, t)
	if p.Sender != 0 {
		out = binary.BigEndian.AppendUint32(out, p.Sender)
	}
	out = binary.BigEndian.AppendUint16(out, uint16(len(p.Payload)))
	return append(out, p.Payload...), nil
}

// New starts a packet of type t; write its fields with W, then Finish it.
func New(t uint16) *Builder { return &Builder{t: t} }

// Builder writes one packet's payload.
type Builder struct {
	t uint16
	W Writer
}

// Finish is the packet with what was written.
func (b *Builder) Finish() *Packet { return &Packet{Type: b.t, Payload: b.W.Bytes()} }
