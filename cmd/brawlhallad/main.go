// Command brawlhallad serves Brawlhalla's own online stack for the Switch build
// (0100C6800B934000, update 10.11).
//
// The game reaches three Blue Mammoth servers, all over its own packet format
// (package netlib):
//
//	user server   TCP. Login, lobby, profile and inventory, matchmaking, custom
//	              rooms. The game dials it by a hard-coded IPv4 address (ports
//	              23001/23002), so consoles are sent here by a signed ip
//	              redirect, not by DNS.
//	game server   UDP with a TCP fallback. A match's input relay. Its address
//	              comes from the user server (GAMESERVER_READY), so it is ours
//	              as soon as the user server is.
//	spectate      the same, for spectators (not served yet).
//
// The wire contract was read out of the Switch build's own executable (its
// symbols are intact), not guessed; docs/protocol.md lists it, and an
// unhandled packet is logged by type and size, because the packet the client
// sends next is the specification for the next handler.
package main

import (
	"flag"
	"log"
	"net"
	"net/http"
)

var (
	userAddr   = flag.String("user-addr", ":22690", "user server TCP listen address")
	healthAddr = flag.String("health-addr", ":22692", "health listen address")
	dumpDir    = flag.String("dump-dir", "", "development only: write every packet but the login's to this directory")
)

func main() {
	flag.Parse()
	srv := newServer(newIdentityClient())

	go func() {
		h := http.NewServeMux()
		h.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"status":"ok"}`))
		})
		log.Printf("[brawlhalla] health on %s", *healthAddr)
		log.Fatal(http.ListenAndServe(*healthAddr, h))
	}()

	ln, err := net.Listen("tcp", *userAddr)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("[brawlhalla] user server on %s", *userAddr)
	for {
		c, err := ln.Accept()
		if err != nil {
			log.Printf("[brawlhalla] accept: %v", err)
			continue
		}
		go srv.serveUser(c)
	}
}
