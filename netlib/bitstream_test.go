package netlib

import (
	"bytes"
	"math"
	"testing"
)

// The widths Packet::SendUnsignedInt picks: bit length rounded up to even,
// behind a 4-bit code of width/2-1.
func TestUnsignedIntWidths(t *testing.T) {
	for _, c := range []struct {
		v    uint32
		bits int
	}{{0, 4 + 2}, {1, 4 + 2}, {3, 4 + 2}, {4, 4 + 4}, {255, 4 + 8}, {256, 4 + 10}, {math.MaxUint32, 4 + 32}} {
		var w Writer
		w.UnsignedInt(c.v)
		if w.Bits() != c.bits {
			t.Errorf("UnsignedInt(%d) wrote %d bits, want %d", c.v, w.Bits(), c.bits)
		}
	}
}

// A byte-aligned 16-bit value is big-endian on the wire: the string length.
func TestStringLayout(t *testing.T) {
	var w Writer
	w.String("Brawlhalla")
	want := append([]byte{0x00, 0x0a}, "Brawlhalla"...)
	if !bytes.Equal(w.Bytes(), want) {
		t.Fatalf("String = % x, want % x", w.Bytes(), want)
	}
}

func TestRoundTrip(t *testing.T) {
	var w Writer
	w.Bool(true)
	w.UnsignedInt(2434)
	w.Int(-77)
	w.Short(-300)
	w.UnsignedShort(65535)
	w.UnsignedInt64(1 << 40)
	w.Float(1.5)
	w.String("héllo")
	w.Byte(0xab)
	w.WriteBits(3, 5)

	r := NewReader(w.Bytes())
	if !r.Bool() || r.UnsignedInt() != 2434 || r.Int() != -77 || r.Short() != -300 ||
		r.UnsignedShort() != 65535 || r.UnsignedInt64() != 1<<40 || r.Float() != 1.5 ||
		r.String() != "héllo" || r.Byte() != 0xab || r.ReadBits(3) != 5 {
		t.Fatal("round trip changed a value")
	}
	if r.Err() != nil {
		t.Fatal(r.Err())
	}
	r.ReadBits(16)
	if r.Err() != ErrShort {
		t.Fatal("reading past the end did not fail")
	}
}
