package game

import (
	"math/rand"

	"github.com/the4ofus/spacetrader-tui/internal/gamedata"
)

type EndGameStatus int

const (
	StatusPlaying EndGameStatus = iota
	StatusRetired
	StatusDead
)

const CurrentSaveVersion = 3

type NewsEntry struct {
	Headline  string `json:"headline"`
	System    string `json:"system"`
	SystemIdx int    `json:"system_idx"`
	Day       int    `json:"day"`
}

type Bookmark struct {
	SystemIdx int    `json:"system_idx"`
	Note      string `json:"note"`
	Day       int    `json:"day"`
}

type GameState struct {
	SaveVersion     int               `json:"save_version"`
	Player          Player            `json:"player"`
	Systems         []SystemState     `json:"systems"`
	CurrentSystemID int               `json:"current_system_id"`
	Day             int               `json:"day"`
	Difficulty      gamedata.Difficulty `json:"difficulty"`
	EndStatus       EndGameStatus     `json:"end_status"`
	Quests          QuestData         `json:"quests"`
	Wormholes       []Wormhole        `json:"wormholes"`
	Seed            int64             `json:"seed"`
	NewsLog         []NewsEntry       `json:"news_log"`
	Bookmarks       []Bookmark        `json:"bookmarks"`

	Rand         *rand.Rand       `json:"-"`
	Data         *gamedata.GameData `json:"-"`
	cargoOfferQty int             `json:"-"`
}

func (gs *GameState) IsBookmarked(sysIdx int) bool {
	for _, b := range gs.Bookmarks {
		if b.SystemIdx == sysIdx {
			return true
		}
	}
	return false
}

func (gs *GameState) GetBookmark(sysIdx int) (Bookmark, bool) {
	for _, b := range gs.Bookmarks {
		if b.SystemIdx == sysIdx {
			return b, true
		}
	}
	return Bookmark{}, false
}

func (gs *GameState) UpdateBookmark(sysIdx int, note string) {
	for i, b := range gs.Bookmarks {
		if b.SystemIdx == sysIdx {
			gs.Bookmarks[i].Note = note
			gs.Bookmarks[i].Day = gs.Day
			return
		}
	}
}

func (gs *GameState) ToggleBookmark(sysIdx int, note string) bool {
	for i, b := range gs.Bookmarks {
		if b.SystemIdx == sysIdx {
			gs.Bookmarks = append(gs.Bookmarks[:i], gs.Bookmarks[i+1:]...)
			return false
		}
	}
	gs.Bookmarks = append(gs.Bookmarks, Bookmark{
		SystemIdx: sysIdx,
		Note:      note,
		Day:       gs.Day,
	})
	return true
}

func (gs *GameState) CurrentSystem() gamedata.SystemDef {
	return gs.Data.Systems[gs.CurrentSystemID]
}

func (gs *GameState) ShipDef(id int) gamedata.ShipDef {
	return gs.Data.Ships[id]
}

func (gs *GameState) EquipDef(id int) gamedata.EquipDef {
	return gs.Data.Equipment[id]
}

func (gs *GameState) PlayerShipDef() gamedata.ShipDef {
	return gs.Data.Ships[gs.Player.Ship.TypeID]
}

func (gs *GameState) EffectiveRange() int {
	shipDef := gs.Data.Ships[gs.Player.Ship.TypeID]
	bonus := 0
	for _, gID := range gs.Player.Ship.Gadgets {
		bonus += gs.Data.Equipment[gID].RangeBonus
	}
	return shipDef.Range + bonus
}
