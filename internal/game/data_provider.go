package game

import "github.com/the4ofus/spacetrader-tui/internal/gamedata"

type ShipDataProvider interface {
	ShipDef(id int) gamedata.ShipDef
	EquipDef(id int) gamedata.EquipDef
}

type GameDataProvider struct {
	Data *gamedata.GameData
}

func (p *GameDataProvider) ShipDef(id int) gamedata.ShipDef {
	return p.Data.Ships[id]
}

func (p *GameDataProvider) EquipDef(id int) gamedata.EquipDef {
	return p.Data.Equipment[id]
}
