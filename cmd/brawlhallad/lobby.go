package main

// BEGIN_LOBBY (2431): the user server's answer to a good login, read by
// LinkUpdater_obj::ReadBeginLobby and the readers it calls. The client reads the
// fields in exactly this order; anything it reads past the end comes back as
// zero, so a short packet does not fail loudly, it just leaves the client with
// the wrong values. Lists are bool-prefixed: true, item, true, item, ..., false.
//
// The first cut sends a fresh player with no history, no playlists and no store
// state. Field names follow what the client does with each value; the ones it only
// stores are named by what is known so far.

import (
	"flag"

	"openpak/brawlhalla/netlib"
)

// regionCount is the length of the client's RegionType::gOrderedRegionTypes, which
// is parsed from the game's own data at boot: BEGIN_LOBBY carries one bool per
// region. Nine is the public region list (US-E, EU, SEA, BRZ, AUS, US-W, JPN, SAF,
// ME); the flag exists until a run confirms it.
var regionCount = flag.Int("regions", 9, "number of region flags in BEGIN_LOBBY (the client's region table length)")

type lobbyState struct {
	Name         string
	BrawlhallaID uint32 // the account id the client keeps (Brawlhalla_obj::sBrawlhallaID)
	XP           uint32
	Level        uint16
	LevelXPStart uint32 // XP at the start of Level
	LevelXPEnd   uint32 // XP at the next level; must differ from LevelXPStart (the client divides by it)
	Tutorials    uint32 // completed-tutorial bits (TutorialManager::ResolveServerDifferences)
	Season       uint32
	SeasonName   string
}

func newLobbyState(p *player) lobbyState {
	id := uint32(p.PID ^ p.PID>>32)
	if id == 0 {
		id = 1
	}
	return lobbyState{
		Name: p.Nickname, BrawlhallaID: id,
		Level: 1, LevelXPStart: 0, LevelXPEnd: 100,
		Season: 1, SeasonName: "OpenPak",
	}
}

func beginLobby(s lobbyState) *netlib.Packet {
	pk := netlib.New(netlib.BeginLobby)
	w := &pk.W
	none := func() { w.Bool(false) } // an empty bool-prefixed list

	w.String(s.Name)
	w.UnsignedInt(s.BrawlhallaID)
	w.Bool(false) // new-user behaviour block absent (would be Bool, UnsignedInt)
	w.String("")  // Game+0x2c8
	w.UnsignedInt(s.XP)
	w.UnsignedShort(s.Level)
	w.UnsignedInt(s.LevelXPStart)
	w.UnsignedInt(s.LevelXPEnd)
	w.UnsignedInt(s.Tutorials)
	w.String("") // Game+0x318
	w.String("") // Game+0x328
	w.String("") // Game+0x338
	w.Byte(0)    // Game+0x26c
	w.Bool(false)
	w.UnsignedInt(0) // Game+0x10
	w.Bool(false)
	w.UnsignedInt(0) // Game+0xc4
	w.UnsignedInt(0) // Game+0xc8
	for i := 0; i < *regionCount; i++ {
		w.Bool(true) // region enabled
	}
	none() // PlaylistType::ReadPlaylistsFromPacket: first list
	none() //   second list
	none() //   optional single playlist
	none() // ReadGameHistory: first list
	none() //   second list
	none() // ReadUserRatingInfo, repeated
	none() // ReadUserPlaylistData
	none() // ReadIgnoreListInfo
	w.Bool(false)
	none() // UnlockManager::ReadStorePrefs: three lists
	none()
	none()
	w.UnsignedInt(s.Season)
	w.String(s.SeasonName)
	w.UnsignedInt(0) // Game+0xac
	w.UnsignedInt(0) // coin sale off
	w.UnsignedInt(0) // active timed event
	w.UnsignedInt(0)
	w.Bool(false)    // BotW mode
	none()           // timed promotions
	w.UnsignedInt(0) // client theme
	w.Byte(1)        // age known: 0 can open the age prompt
	w.UnsignedInt(0)
	w.UnsignedInt(0) // Game+0x134
	w.UnsignedInt(0) // four values into Game+0x138 (+0x48..+0x54); the second picks the notice text
	w.UnsignedInt(0)
	w.UnsignedInt(0)
	w.UnsignedInt(0)
	w.String("")  // Game+0x348
	w.Bool(false) // server notification block absent (UnsignedInt, UnsignedInt, String, optional String)
	return pk.Finish()
}
