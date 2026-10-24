# brawlhalla - OpenPak's Brawlhalla backend

The services the Switch build of Brawlhalla (`0100C6800B934000`, update 10.11) reaches when a
console is redirected to OpenPak.

Brawlhalla talks to three servers of Blue Mammoth's own, all in its own packet format (`netlib/`):

| Server | Transport | What it does |
| --- | --- | --- |
| user server | TCP | login, lobby, profile and inventory, matchmaking, custom rooms |
| game server | UDP, TCP fallback | a match's input relay (the clients simulate the match themselves, with rollback) |
| spectate server | UDP/TCP | spectators |

The game dials the user server by a **hard-coded IPv4 address**, not a name, so DNS redirection
cannot reach it. Consoles and Harbor get there through a signed ip redirect: the ceiling's
`ip_redirects` section, applied by Harbor's `bsd` service and by the `bsd:u` MITM in OpenPak's boot
package. The user server sends the game server's and spectate server's addresses itself.

## Status: early

- Packet codec (bitstream, framing, packet types): done, with tests.
- User-server login (handshake, version, challenge, Nintendo token check through nx-baas): written,
  not yet run against the game.
- Lobby (`BEGIN_LOBBY`), matchmaking, custom rooms, game server: not started.

Store, currency and unlocks are planned as "everything owned", served locally, with no purchases.

## Ports

| Port | |
| --- | --- |
| 22690/tcp | user server |
| 22691/udp+tcp | game server (planned) |
| 22692/tcp | health (`/health`) |
| 22693/udp+tcp | spectate server (planned) |

## Running it

```
go run ./cmd/brawlhallad -user-addr :22690 -health-addr :22692
```

`NX_INTERNAL_URL` and `NX_INTERNAL_KEY` point it at nx-baas for the token check.

The wire contract is in [docs/protocol.md](docs/protocol.md). It was read out of the game client's
own executable. No Blue Mammoth server code is used.

## Licence

AGPL-3.0-only.
