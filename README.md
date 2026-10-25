# Brawlhalla

OpenPak's Brawlhalla Server.

| | |
| --- | --- |
| Game | [Brawlhalla](https://openpak.org/titles#t-64) |
| Network | Nintendo Switch |
| Title ID | `0100C6800B934000` |
| Licence | AGPL-3.0-only |

## What it serves

| Server | Transport | What it does | Served by |
| --- | --- | --- | --- |
| User server | TCP | login, lobby, profile and inventory, matchmaking, custom rooms | `brawlhallad` |
| Game server | UDP, TCP fallback | a match's input relay | — |
| Spectate server | UDP/TCP | spectators | — |

The game reaches the user server by a hard-coded IPv4 address. See [protocol.md](docs/protocol.md).

## Run

```sh
go run ./cmd/brawlhallad -user-addr :22690 -health-addr :22692
```

| Setting | Default | Meaning |
| --- | --- | --- |
| `-user-addr` | `:22690` | user server TCP listen address |
| `-health-addr` | `:22692` | health listen address |
| `NX_INTERNAL_URL` | — | nx-baas address for the Nintendo token check |
| `NX_INTERNAL_KEY` | — | key for nx-baas |

## Docs

- [Protocol](docs/protocol.md)
- [Evidence](docs/evidence.md)
- [Deploy](docs/deploy.md)
- [Notes](docs/notes.md)

Live compatibility status: [openpak.org/titles#t-64](https://openpak.org/titles#t-64).

## Licence

AGPL-3.0-only. See [LICENSE](LICENSE).
