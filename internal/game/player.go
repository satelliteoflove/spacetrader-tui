package game

import "github.com/the4ofus/spacetrader-tui/internal/formula"

type Player struct {
	Name          string          `json:"name"`
	Credits       int             `json:"credits"`
	LoanBalance   int             `json:"loan_balance"`
	Skills        [formula.NumSkills]int `json:"skills"`
	PoliceRecord  int             `json:"police_record"`
	Reputation    int             `json:"reputation"`
	Ship          Ship            `json:"ship"`
	Cargo         [10]int         `json:"cargo"`
	CargoCost     [10]int         `json:"cargo_cost"`
	Crew          []Mercenary     `json:"crew"`
	HasEscapePod  bool            `json:"has_escape_pod"`
	HasInsurance  bool            `json:"has_insurance"`
	InsuranceDays int             `json:"insurance_days"`
	MoonPurchased bool            `json:"moon_purchased"`
}

func (p *Player) TotalCargo() int {
	total := 0
	for _, qty := range p.Cargo {
		total += qty
	}
	return total
}

func (p *Player) CargoCapacity(data ShipDataProvider) int {
	def := data.ShipDef(p.Ship.TypeID)
	extra := 0
	for _, g := range p.Ship.Gadgets {
		eq := data.EquipDef(g)
		extra += eq.CargoBays
	}
	return def.CargoBays + extra
}

func (p *Player) FreeCargo(data ShipDataProvider) int {
	return p.CargoCapacity(data) - p.TotalCargo()
}

func (p *Player) CrewMercs() []formula.Mercenary {
	mercs := make([]formula.Mercenary, len(p.Crew))
	for i, m := range p.Crew {
		mercs[i] = m
	}
	return mercs
}
