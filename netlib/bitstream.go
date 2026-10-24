// Package netlib speaks Blue Mammoth's packet format: the bitstream a Brawlhalla
// packet's payload is written in, and (framing.go) the packets themselves.
//
// The layout was read out of the Switch build's own executable (its native
// BitStream and Packet classes), not guessed; docs/protocol.md lists what each
// field type looks like on the wire. Bits are written most significant first,
// so a value that starts on a byte boundary reads like big-endian.
package netlib

import (
	"errors"
	"math"
	"math/bits"
)

// ErrShort is returned when a read runs past the end of the payload.
var ErrShort = errors.New("netlib: read past the end of the packet")

// Writer builds a payload.
type Writer struct {
	buf  []byte
	nbit int
}

// Bytes is the payload so far, the last byte zero-padded.
func (w *Writer) Bytes() []byte { return w.buf }

// Bits is how many bits have been written.
func (w *Writer) Bits() int { return w.nbit }

// WriteBits writes the low n bits of v (n <= 64), most significant first.
func (w *Writer) WriteBits(n int, v uint64) {
	for i := n - 1; i >= 0; i-- {
		if w.nbit%8 == 0 {
			w.buf = append(w.buf, 0)
		}
		if v>>uint(i)&1 != 0 {
			w.buf[w.nbit/8] |= 0x80 >> uint(w.nbit%8)
		}
		w.nbit++
	}
}

// Bool is one bit.
func (w *Writer) Bool(b bool) {
	if b {
		w.WriteBits(1, 1)
	} else {
		w.WriteBits(1, 0)
	}
}

// Byte is eight bits.
func (w *Writer) Byte(b byte) { w.WriteBits(8, uint64(b)) }

// Bytes writes raw bytes at the current bit position.
func (w *Writer) Raw(p []byte) {
	for _, b := range p {
		w.Byte(b)
	}
}

// packedLen is the field width the packed integers use: the value's bit length
// (at least 1), rounded up to an even number.
func packedLen(v uint64) int {
	n := bits.Len64(v)
	if n == 0 {
		n = 1
	}
	return n + n&1
}

// UnsignedInt is Packet::SendUnsignedInt: a 4-bit width code (width/2 - 1),
// then the value in that many bits.
func (w *Writer) UnsignedInt(v uint32) {
	n := packedLen(uint64(v))
	w.WriteBits(4, uint64(n/2-1))
	w.WriteBits(n, uint64(v))
}

// Int is Packet::SendInt: a sign bit, then the magnitude as UnsignedInt.
func (w *Writer) Int(v int32) {
	m := uint32(v)
	if v < 0 {
		w.WriteBits(1, 1)
		m = uint32(-int64(v))
	} else {
		w.WriteBits(1, 0)
	}
	w.UnsignedInt(m)
}

// UnsignedShort is Packet::SendUnsignedShort: a 3-bit width code, then the value.
func (w *Writer) UnsignedShort(v uint16) {
	n := packedLen(uint64(v))
	w.WriteBits(3, uint64(n/2-1))
	w.WriteBits(n, uint64(v))
}

// Short is Packet::SendShort: a sign bit, then the magnitude as UnsignedShort.
func (w *Writer) Short(v int16) {
	m := uint16(v)
	if v < 0 {
		w.WriteBits(1, 1)
		m = uint16(-int32(v))
	} else {
		w.WriteBits(1, 0)
	}
	w.UnsignedShort(m)
}

// UnsignedInt64 is Packet::SendUnsignedInt64: a 5-bit width code, then the value.
func (w *Writer) UnsignedInt64(v uint64) {
	n := packedLen(v)
	w.WriteBits(5, uint64(n/2-1))
	w.WriteBits(n, v)
}

// Float is Packet::SendFloat: the IEEE 754 bits, most significant first.
func (w *Writer) Float(f float32) { w.WriteBits(32, uint64(math.Float32bits(f))) }

// String is Packet::SendString: a 16-bit length, then the bytes (no terminator).
func (w *Writer) String(s string) {
	if len(s) > 0xffff {
		s = s[:0xffff]
	}
	w.WriteBits(16, uint64(len(s)))
	w.Raw([]byte(s))
}

// Reader takes a payload apart.
type Reader struct {
	buf  []byte
	nbit int
	err  error
}

// NewReader reads p from its first bit.
func NewReader(p []byte) *Reader { return &Reader{buf: p} }

// Err is the first error a read met; every read after it returns zero.
func (r *Reader) Err() error { return r.err }

// Remaining is how many bits are left.
func (r *Reader) Remaining() int { return len(r.buf)*8 - r.nbit }

// ReadBits reads n bits (n <= 64), most significant first.
func (r *Reader) ReadBits(n int) uint64 {
	if r.err != nil {
		return 0
	}
	if n > r.Remaining() {
		r.err = ErrShort
		return 0
	}
	var v uint64
	for i := 0; i < n; i++ {
		v <<= 1
		if r.buf[r.nbit/8]&(0x80>>uint(r.nbit%8)) != 0 {
			v |= 1
		}
		r.nbit++
	}
	return v
}

func (r *Reader) Bool() bool          { return r.ReadBits(1) == 1 }
func (r *Reader) Byte() byte          { return byte(r.ReadBits(8)) }
func (r *Reader) Float() float32      { return math.Float32frombits(uint32(r.ReadBits(32))) }
func (r *Reader) UnsignedInt() uint32 { return uint32(r.ReadBits(int(r.ReadBits(4)+1) * 2)) }
func (r *Reader) UnsignedShort() uint16 {
	return uint16(r.ReadBits(int(r.ReadBits(3)+1) * 2))
}
func (r *Reader) UnsignedInt64() uint64 { return r.ReadBits(int(r.ReadBits(5)+1) * 2) }

func (r *Reader) Int() int32 {
	neg := r.Bool()
	m := int64(r.UnsignedInt())
	if neg {
		m = -m
	}
	return int32(m)
}

func (r *Reader) Short() int16 {
	neg := r.Bool()
	m := int32(r.UnsignedShort())
	if neg {
		m = -m
	}
	return int16(m)
}

// Raw reads n bytes.
func (r *Reader) Raw(n int) []byte {
	if n*8 > r.Remaining() {
		r.err = ErrShort
		return nil
	}
	out := make([]byte, n)
	for i := range out {
		out[i] = r.Byte()
	}
	return out
}

// String reads a 16-bit length and that many bytes.
func (r *Reader) String() string { return string(r.Raw(int(r.ReadBits(16)))) }
