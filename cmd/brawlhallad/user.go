package main

// The user server: one TCP connection per signed-in player.
//
// Login, as the client drives it (LinkUpdater_obj, Game_obj::HandleFirstLogin,
// LoginManager_obj::CreateAuthPacket):
//
//	client  CLIENT_SERVER_TCP_HANDSHAKE (178)  String "Brawlhalla client to server protocol 1.0"
//	client  LOGIN_VERSION (30)                 UnsignedInt client version (10110 in 10.11), UnsignedInt
//	server  LOGIN_CHALLENGE (12000)            String challenge
//	client  LOGIN_AUTHENTICATE_SWITCH (131)    String, UnsignedInt n, n bytes: the Nintendo id
//	                                           token XORed with the challenge (GameUtil::EncryptString,
//	                                           key[(offset+i) % len(key)]), UnsignedInt 26, checksum info
//	server  BEGIN_LOBBY (2431) ...
//
// The challenge is one repeated character, so the XOR does not depend on the
// offset the client starts the key at.

import (
	"bufio"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"sync/atomic"
	"time"

	"openpak/brawlhalla/netlib"
	"openpak/brawlhalla/redact"
)

const handshakeBanner = "Brawlhalla client to server protocol 1.0"

type server struct {
	ids identityClient
	seq atomic.Uint64
}

func newServer(ids identityClient) *server { return &server{ids: ids} }

type userConn struct {
	srv       *server
	c         net.Conn
	w         *bufio.Writer
	tag       string
	n         uint64
	challenge byte // 0 until LOGIN_CHALLENGE went out
	version   uint32
	player    *player
}

func (s *server) serveUser(c net.Conn) {
	defer c.Close()
	u := &userConn{srv: s, c: c, w: bufio.NewWriter(c), n: s.seq.Add(1)}
	u.tag = redact.Tag(c.RemoteAddr().String())
	log.Printf("[user %s] connected", u.tag)
	r := bufio.NewReader(c)
	for {
		c.SetReadDeadline(time.Now().Add(2 * time.Minute)) // the client pings every 10 s
		p, err := netlib.ReadPacket(r)
		if err != nil {
			if !errors.Is(err, io.EOF) {
				log.Printf("[user %s] read: %s", u.tag, redact.Err(err))
			}
			log.Printf("[user %s] closed", u.tag)
			return
		}
		u.dump("in", p)
		if err := u.handle(p); err != nil {
			log.Printf("[user %s] packet %d: %v; closing", u.tag, p.Type, err)
			return
		}
		if err := u.w.Flush(); err != nil {
			log.Printf("[user %s] write: %s", u.tag, redact.Err(err))
			return
		}
	}
}

func (u *userConn) send(p *netlib.Packet) error {
	raw, err := p.Encode()
	if err != nil {
		return err
	}
	u.dump("out", p)
	_, err = u.w.Write(raw)
	return err
}

func (u *userConn) handle(p *netlib.Packet) error {
	r := netlib.NewReader(p.Payload)
	switch p.Type {
	case netlib.ClientServerTcpHandshake:
		if banner := r.String(); banner != handshakeBanner {
			return fmt.Errorf("unknown handshake %q", banner)
		}
		return nil
	case netlib.LoginVersion:
		u.version = r.UnsignedInt()
		extra := r.UnsignedInt()
		log.Printf("[user %s] client version %d (%d)", u.tag, u.version, extra)
		return u.sendChallenge()
	case netlib.LoginAuthenticateSwitch:
		return u.login(r)
	case netlib.ClientUserserverPing:
		return u.send(netlib.New(netlib.ClientUserserverPing).Finish())
	}
	log.Printf("[user %s] unhandled packet %d (%d bytes)", u.tag, p.Type, len(p.Payload))
	return nil
}

func (u *userConn) sendChallenge() error {
	var b [1]byte
	rand.Read(b[:])
	u.challenge = 'A' + b[0]%26
	key := make([]byte, 32)
	for i := range key {
		key[i] = u.challenge
	}
	pk := netlib.New(netlib.LoginChallenge)
	pk.W.String(string(key))
	return u.send(pk.Finish())
}

func (u *userConn) login(r *netlib.Reader) error {
	first := r.String()
	n := r.UnsignedInt()
	if n > 8192 {
		return fmt.Errorf("token of %d bytes", n)
	}
	tok := r.Raw(int(n))
	for i := range tok {
		tok[i] ^= u.challenge
	}
	after := r.UnsignedInt()
	if r.Err() != nil {
		return r.Err()
	}
	log.Printf("[user %s] switch login: first string %d bytes, token %d bytes, then %d, %d bits of checksum info",
		u.tag, len(first), len(tok), after, r.Remaining())
	p, err := u.srv.ids.byIDToken(string(tok))
	if err != nil {
		log.Printf("[user %s] login refused: %s", u.tag, redact.Err(err))
		return u.send(netlib.New(netlib.SwitchNoAuth).Finish())
	}
	u.player = p
	log.Printf("[user %s] signed in as %s", u.tag, redact.Tag(p.PID))
	return nil // BEGIN_LOBBY next: see docs/protocol.md
}

// dump writes a packet to -dump-dir (development only). The login packet carries
// the console's token and is never written.
func (u *userConn) dump(dir string, p *netlib.Packet) {
	if *dumpDir == "" || p.Type == netlib.LoginAuthenticateSwitch {
		return
	}
	name := fmt.Sprintf("%d-%s-%d-%s.bin", u.n, time.Now().Format("150405.000"), p.Type, dir)
	os.WriteFile(filepath.Join(*dumpDir, name), p.Payload, 0o600)
}
