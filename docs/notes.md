# Notes: brawlhalla

## Previous title

brawlhalla - OpenPak's Brawlhalla backend

## Status: early

- Packet codec (bitstream, framing, packet types): done, with tests.
- User-server login (handshake, version, challenge, Nintendo token check through nx-baas): written,
  not yet run against the game.
- Lobby (`BEGIN_LOBBY`), matchmaking, custom rooms, game server: not started.

Store, currency and unlocks are planned as "everything owned", served locally, with no purchases.

## Licence

AGPL-3.0-only.
