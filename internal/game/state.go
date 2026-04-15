package game

import (
	"math/rand"

	"github.com/the4ofus/spacetrader-tui/internal/formula"
	"github.com/the4ofus/spacetrader-tui/internal/gamedata"
)

type EndGameStatus int

const (
	StatusPlaying EndGameStatus = iota
	StatusRetired
	StatusDead
)

const CurrentSaveVersion = 6

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

type TradeInfoEntry struct {
	Prices [NumGoods]int `json:"prices"`
	Event  string        `json:"event"`
	Day    int           `json:"day"`
}

type DailySnapshot struct {
	Day      int `json:"day"`
	Credits  int `json:"credits"`
	NetWorth int `json:"net_worth"`
}

const TradeInfoStaleDays = 5
const TradeInfoBuyCost = 100
const TradeInfoMaxRange = 15.0

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
	ActiveRoute       int             `json:"active_route"`
	ActiveRouteOrigin int            `json:"active_route_origin"`
	HasActiveRoute    bool           `json:"has_active_route"`
	Mercenaries       []Mercenary              `json:"mercenaries"`
	TradeInfo         map[int]TradeInfoEntry   `json:"trade_info,omitempty"`
	Ledger            []DailySnapshot          `json:"ledger,omitempty"`

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

func (gs *GameState) RecordSnapshot() {
	dp := &GameDataProvider{Data: gs.Data}
	gs.Ledger = append(gs.Ledger, DailySnapshot{
		Day:      gs.Day,
		Credits:  gs.Player.Credits,
		NetWorth: gs.Player.Worth(dp),
	})
}

func (gs *GameState) CaptureTradeInfo(sysIdx int) {
	if gs.TradeInfo == nil {
		gs.TradeInfo = make(map[int]TradeInfoEntry)
	}
	gs.TradeInfo[sysIdx] = TradeInfoEntry{
		Prices: gs.Systems[sysIdx].Prices,
		Event:  gs.Systems[sysIdx].Event,
		Day:    gs.Day,
	}
}

func (gs *GameState) GetTradeInfo(sysIdx int) (TradeInfoEntry, bool) {
	if gs.TradeInfo == nil {
		return TradeInfoEntry{}, false
	}
	info, ok := gs.TradeInfo[sysIdx]
	return info, ok
}

func (gs *GameState) IsTradeInfoStale(sysIdx int) (bool, int) {
	info, ok := gs.GetTradeInfo(sysIdx)
	if !ok {
		return true, 0
	}
	age := gs.Day - info.Day
	if age >= TradeInfoStaleDays {
		return true, age
	}
	return false, age
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

func (gs *GameState) HopsToSystem(destIdx int) int {
	systems := make([][2]int, len(gs.Data.Systems))
	for i, sys := range gs.Data.Systems {
		systems[i] = [2]int{sys.X, sys.Y}
	}
	wormholes := make([]formula.WormholePair, len(gs.Wormholes))
	for i, wh := range gs.Wormholes {
		wormholes[i] = formula.WormholePair{A: wh.SystemA, B: wh.SystemB}
	}
	return formula.ShortestPathHops(systems, gs.EffectiveRange(), wormholes, gs.CurrentSystemID, destIdx)
}
