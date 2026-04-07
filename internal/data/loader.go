package data

import (
	"encoding/json"
	"fmt"
	"io/fs"

	"github.com/the4ofus/spacetrader-tui/internal/gamedata"
)

func LoadAll(dataFS fs.FS) (*gamedata.GameData, error) {
	gd := &gamedata.GameData{}
	var err error

	gd.Goods, err = loadGoods(dataFS)
	if err != nil {
		return nil, fmt.Errorf("loading goods: %w", err)
	}

	gd.Ships, err = loadShips(dataFS)
	if err != nil {
		return nil, fmt.Errorf("loading ships: %w", err)
	}

	gd.Equipment, err = loadEquipment(dataFS)
	if err != nil {
		return nil, fmt.Errorf("loading equipment: %w", err)
	}

	return gd, nil
}

func LoadAllFromEmbed(dataFS fs.FS) (*gamedata.GameData, error) {
	sub, err := fs.Sub(dataFS, "data")
	if err != nil {
		return nil, fmt.Errorf("getting data subdirectory: %w", err)
	}
	return LoadAll(sub)
}

type rawGood struct {
	Name               string `json:"name"`
	BasePrice          int    `json:"base_price"`
	Legality           bool   `json:"legality"`
	MinTech            string `json:"min_tech"`
	MaxTech            string `json:"max_tech"`
	Variance           int    `json:"variance"`
	PriceIncreaseEvent string `json:"price_increase_event"`
	PriceDecreaseEvent string `json:"price_decrease_event"`
	ExpensiveResource  string `json:"expensive_resource"`
	CheapResource      string `json:"cheap_resource"`
}

func loadGoods(dataFS fs.FS) ([]gamedata.GoodDef, error) {
	raw, err := readJSON[[]rawGood](dataFS, "goods.json")
	if err != nil {
		return nil, err
	}

	goods := make([]gamedata.GoodDef, len(raw))
	for i, r := range raw {
		minTech, err := gamedata.ParseTechLevel(r.MinTech)
		if err != nil {
			return nil, fmt.Errorf("good %q: %w", r.Name, err)
		}
		maxTech, err := gamedata.ParseTechLevel(r.MaxTech)
		if err != nil {
			return nil, fmt.Errorf("good %q: %w", r.Name, err)
		}
		id, err := gamedata.ParseGoodID(r.Name)
		if err != nil {
			return nil, fmt.Errorf("good %q: %w", r.Name, err)
		}
		goods[i] = gamedata.GoodDef{
			ID:                 id,
			Name:               r.Name,
			BasePrice:          r.BasePrice,
			Legal:              r.Legality,
			MinTech:            minTech,
			MaxTech:            maxTech,
			Variance:           r.Variance,
			PriceIncreaseEvent: r.PriceIncreaseEvent,
			PriceDecreaseEvent: r.PriceDecreaseEvent,
			ExpensiveResource:  r.ExpensiveResource,
			CheapResource:      r.CheapResource,
		}
	}
	return goods, nil
}

type rawShip struct {
	Type         string `json:"type"`
	Size         string `json:"size"`
	CargoBays    int    `json:"cargo_bays"`
	WeaponSlots  int    `json:"weapon_slots"`
	ShieldSlots  int    `json:"shield_slots"`
	GadgetSlots  int    `json:"gadget_slots"`
	CrewQuarters int    `json:"crew_quarters"`
	Range        int    `json:"range"`
	FuelCost     int    `json:"fuel_cost"`
	Hull         int    `json:"hull"`
	RepairCost   int    `json:"repair_cost"`
	Price        int    `json:"price"`
	MinTech      string `json:"min_tech"`
}

func loadShips(dataFS fs.FS) ([]gamedata.ShipDef, error) {
	raw, err := readJSON[[]rawShip](dataFS, "ships.json")
	if err != nil {
		return nil, err
	}

	ships := make([]gamedata.ShipDef, len(raw))
	for i, r := range raw {
		size, err := gamedata.ParseShipSize(r.Size)
		if err != nil {
			return nil, fmt.Errorf("ship %q: %w", r.Type, err)
		}
		minTech, err := gamedata.ParseTechLevel(r.MinTech)
		if err != nil {
			return nil, fmt.Errorf("ship %q: %w", r.Type, err)
		}
		ships[i] = gamedata.ShipDef{
			ID:           i,
			Name:         r.Type,
			Size:         size,
			CargoBays:    r.CargoBays,
			WeaponSlots:  r.WeaponSlots,
			ShieldSlots:  r.ShieldSlots,
			GadgetSlots:  r.GadgetSlots,
			CrewQuarters: r.CrewQuarters,
			Range:        r.Range,
			FuelCost:     r.FuelCost,
			Hull:         r.Hull,
			RepairCost:   r.RepairCost,
			Price:        r.Price,
			MinTech:      minTech,
		}
	}
	return ships, nil
}

type rawEquipment struct {
	Name      string      `json:"name"`
	Type      string      `json:"type"`
	Power     int         `json:"power"`
	Protect   int         `json:"protection"`
	Bonus     interface{} `json:"bonus"`
	TechLevel string      `json:"tech_level"`
	Price     int         `json:"price"`
}

func loadEquipment(dataFS fs.FS) ([]gamedata.EquipDef, error) {
	raw, err := readJSON[[]rawEquipment](dataFS, "equipment.json")
	if err != nil {
		return nil, err
	}

	equipment := make([]gamedata.EquipDef, len(raw))
	for i, r := range raw {
		cat, err := gamedata.ParseEquipCategory(r.Type)
		if err != nil {
			return nil, fmt.Errorf("equipment %q: %w", r.Name, err)
		}
		tech, err := gamedata.ParseTechLevel(r.TechLevel)
		if err != nil {
			return nil, fmt.Errorf("equipment %q: %w", r.Name, err)
		}

		eq := gamedata.EquipDef{
			ID:        i,
			Name:      r.Name,
			Category:  cat,
			Power:     r.Power,
			Protection: r.Protect,
			TechLevel: tech,
			Price:     r.Price,
		}

		switch b := r.Bonus.(type) {
		case float64:
			eq.CargoBays = int(b)
		case string:
			eq.SkillBonus = b
		}

		equipment[i] = eq
	}
	equipment = append(equipment, questRewardEquipment(len(equipment))...)
	return equipment, nil
}

func questRewardEquipment(startID int) []gamedata.EquipDef {
	return []gamedata.EquipDef{
		{
			ID:          startID,
			Name:        "Morgan's Laser",
			Category:    gamedata.EquipWeapon,
			Power:       85,
			TechLevel:   gamedata.TechHiTech,
			Price:       50000,
			QuestReward: true,
		},
		{
			ID:          startID + 1,
			Name:        "Lightning Shield",
			Category:    gamedata.EquipShield,
			Protection:  350,
			TechLevel:   gamedata.TechHiTech,
			Price:       45000,
			QuestReward: true,
		},
		{
			ID:          startID + 2,
			Name:        "Fuel Compactor",
			Category:    gamedata.EquipGadget,
			RangeBonus:  3,
			TechLevel:   gamedata.TechHiTech,
			Price:       30000,
			QuestReward: true,
		},
	}
}

func readJSON[T any](dataFS fs.FS, filename string) (T, error) {
	var result T
	b, err := fs.ReadFile(dataFS, filename)
	if err != nil {
		return result, fmt.Errorf("reading %s: %w", filename, err)
	}
	if err := json.Unmarshal(b, &result); err != nil {
		return result, fmt.Errorf("parsing %s: %w", filename, err)
	}
	return result, nil
}
