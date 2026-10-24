# Brawlhalla wire protocol (Switch 10.11)

Read out of the Switch client's own executable. The client's names (`PKTTYPE_*`, the
`LinkUpdater` read handlers) are kept, so a field here can be found in the client again.

## Packets on TCP

```
u16 type      big-endian; bit 15 set means a u32 sender follows
[u32 sender]  only for GAMESERVER_LOGIN (10400), GAMESERVER_LOGIN_REJOIN (10405), CONNECT_SPECTATE (10202)
u16 length    payload bytes
payload       a bitstream
```

Packet type ranges: client to user server 30-194, user server to client 2401-2954, spectate
10100-10202, game server to client 10300-10316, client to game server 10400-10423, login
12000-12002, ping 12100. The full list is in `netlib/types.go`.

## Bitstream

Bits are written most significant first. A value that starts on a byte boundary therefore reads
like big-endian.

| Field | Encoding |
| --- | --- |
| Bool | 1 bit |
| Byte | 8 bits |
| UnsignedInt | 4-bit code c, then 2(c+1) bits: the bit length rounded up to even, at least 2 |
| Int | sign bit, then the magnitude as UnsignedInt |
| UnsignedShort | 3-bit code, then 2(c+1) bits |
| Short | sign bit, then the magnitude as UnsignedShort |
| UnsignedInt64 | 5-bit code, then 2(c+1) bits |
| Float | 32 IEEE 754 bits |
| String | 16-bit length, then the bytes, with no terminator |

## User-server login

| Direction | Packet | Fields |
| --- | --- | --- |
| client | CLIENT_SERVER_TCP_HANDSHAKE (178) | String `Brawlhalla client to server protocol 1.0` |
| client | LOGIN_VERSION (30) | UnsignedInt version (10110), UnsignedInt |
| server | LOGIN_CHALLENGE (12000) | String challenge |
| client | LOGIN_AUTHENTICATE_SWITCH (131) | String, UnsignedInt n, n bytes, UnsignedInt 26, checksum info |
| server | BEGIN_LOBBY (2431), or SWITCH_NO_AUTH (12002) | profile, region flags, playlists, history, store prefs, season; field order in `cmd/brawlhallad/lobby.go` |
| client | CLIENT_USERSERVER_PING (12100) | every 10 s; the server answers with the same type |

In the login packet, the n bytes are the console's Nintendo id token XORed with the challenge:
byte i uses `key[(offset + i) % len(key)]`. The server sends a challenge made of one repeated
character, so the starting offset does not matter.

## Addresses

The client picks its user server from a small table of IPv4 addresses compiled into it, on TCP port
23001 or 23002. The only override is a command-line flag that a console never has, so OpenPak
redirects the address itself (a signed `ip_redirects` rule). The game server's address arrives in
GAMESERVER_READY (2434).

## BEGIN_LOBBY

Read by `LinkUpdater::ReadBeginLobby`. Lists are bool-prefixed (true, item, ..., false).
The packet carries one bool per entry of the client's region table, which the client loads from
its own data files, so the server has to send exactly that many (`-regions`, 9 by default).
The client divides by (level-end XP - level-start XP), so the two must differ, and an age byte
of 0 can open the age prompt. The full order is in `cmd/brawlhallad/lobby.go`.
