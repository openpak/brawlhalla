# Deploy: brawlhalla

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

## Build

The repo's `Dockerfile` builds `./cmd/brawlhallad` into a `gcr.io/distroless/static` image that runs as `nonroot:nonroot` and exposes ports 22690 and 22692. The image name is not recorded in the README or in CI.

## Flags

| Flag | Default | Meaning |
| --- | --- | --- |
| `-user-addr` | `:22690` | user server TCP listen address |
| `-health-addr` | `:22692` | health listen address |
| `-regions` | `9` | number of region flags in BEGIN_LOBBY |
| `-dump-dir` | empty | development only: write every packet but the login's to this directory |
| `-dev-trust-tokens` | `false` | development only: accept any login token without asking nx-baas |
