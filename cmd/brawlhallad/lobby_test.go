package main

import (
	"testing"

	"openpak/brawlhalla/netlib"
)

// TestBeginLobbyOrder reads the packet back in ReadBeginLobby's order and checks
// that every bit is consumed.
func TestBeginLobbyOrder(t *testing.T) {
	s := newLobbyState(&player{PID: 0x1234_5678_9abc, Nickname: "tester"})
	r := netlib.NewReader(beginLobby(s).Payload)
	if got := r.String(); got != "tester" {
		t.Fatalf("name %q", got)
	}
	if got := r.UnsignedInt(); got != s.BrawlhallaID || got == 0 {
		t.Fatalf("id %d", got)
	}
	r.Bool()
	_ = r.String()
	r.UnsignedInt()
	if got := r.UnsignedShort(); got != 1 {
		t.Fatalf("level %d", got)
	}
	if a, b := r.UnsignedInt(), r.UnsignedInt(); a == b {
		t.Fatal("level XP range is empty")
	}
	r.UnsignedInt()
	_ = r.String()
	_ = r.String()
	_ = r.String()
	r.Byte()
	r.Bool()
	r.UnsignedInt()
	r.Bool()
	r.UnsignedInt()
	r.UnsignedInt()
	for i := 0; i < *regionCount; i++ {
		if !r.Bool() {
			t.Fatalf("region %d off", i)
		}
	}
	for i := 0; i < 8; i++ { // playlists 3, history 2, ratings, playlist data, ignore list
		if r.Bool() {
			t.Fatalf("list %d not empty", i)
		}
	}
	r.Bool()
	r.Bool()
	r.Bool()
	r.Bool()
	r.UnsignedInt()
	if got := r.String(); got != s.SeasonName {
		t.Fatalf("season %q", got)
	}
	r.UnsignedInt()
	r.UnsignedInt()
	r.UnsignedInt()
	r.UnsignedInt()
	r.Bool()
	r.Bool()
	r.UnsignedInt()
	if got := r.Byte(); got == 0 {
		t.Fatal("age byte 0 opens the age prompt")
	}
	for i := 0; i < 6; i++ {
		r.UnsignedInt()
	}
	_ = r.String()
	r.Bool()
	if r.Err() != nil || r.Remaining() >= 8 {
		t.Fatalf("err %v, %d bits left", r.Err(), r.Remaining())
	}
}
