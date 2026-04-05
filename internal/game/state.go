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

type GameState struct {
	Player          Player            `json:"player"`
	Systems         []SystemState     `json:"systems"`
	CurrentSystemID int               `json:"current_system_id"`
	Day             int               `json:"day"`
	Difficulty      gamedata.Difficulty `json:"difficulty"`
	EndStatus       EndGameStatus     `json:"end_status"`
	Quests          QuestData         `json:"quests"`
	Wormholes       []Wormhole        `json:"wormholes"`
	Seed            int64             `json:"seed"`

	Rand *rand.Rand       `json:"-"`
	Data *gamedata.GameData `json:"-"`
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
