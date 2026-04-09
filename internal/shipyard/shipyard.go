package shipyard

import (
	"fmt"

	"github.com/the4ofus/spacetrader-tui/internal/game"
	"github.com/the4ofus/spacetrader-tui/internal/gamedata"
)

type Result struct {
	Success bool
	Message string
}

func AvailableShips(gs *game.GameState) []gamedata.ShipDef {
	sys := gs.Data.Systems[gs.CurrentSystemID]
	var ships []gamedata.ShipDef
	for _, s := range gs.Data.Ships {
		if s.MinTech <= sys.TechLevel {
			ships = append(ships, s)
		}
	}
	return ships
}

func AvailableEquipment(gs *game.GameState) []gamedata.EquipDef {
	sys := gs.Data.Systems[gs.CurrentSystemID]
	var equip []gamedata.EquipDef
	for _, e := range gs.Data.Equipment {
		if e.TechLevel <= sys.TechLevel && !e.QuestReward {
			equip = append(equip, e)
		}
	}
	return equip
}

func TradeInValue(gs *game.GameState) int {
	shipDef := gs.Data.Ships[gs.Player.Ship.TypeID]

	shipMultNum := 3
	if gs.Quests.TribbleQty > 0 {
		shipMultNum = 1
	}
	value := shipDef.Price * shipMultNum / 4

	for _, wID := range gs.Player.Ship.Weapons {
		value += gs.Data.Equipment[wID].Price * 2 / 3
	}
	for _, sID := range gs.Player.Ship.Shields {
		value += gs.Data.Equipment[sID].Price * 2 / 3
	}
	for _, gID := range gs.Player.Ship.Gadgets {
		value += gs.Data.Equipment[gID].Price * 2 / 3
	}

	return value
}

func BuyShip(gs *game.GameState, shipTypeID int) Result {
	if shipTypeID < 0 || shipTypeID >= len(gs.Data.Ships) {
		return Result{Message: "Invalid ship type."}
	}
	if game.ReactorOnBoard(gs) {
		return Result{Message: "Can't trade ships while carrying the Ion Reactor. Deliver it first."}
	}

	newShip := gs.Data.Ships[shipTypeID]
	sys := gs.Data.Systems[gs.CurrentSystemID]
	if newShip.MinTech > sys.TechLevel {
		return Result{Message: "Ship not available at this tech level."}
	}

	tradeIn := TradeInValue(gs)
	cost := newShip.Price - tradeIn

	if cost > gs.Player.Credits {
		return Result{Message: fmt.Sprintf("Not enough credits. Need %d more.", cost-gs.Player.Credits)}
	}

	cargoCount := gs.Player.TotalCargo()
	if cargoCount > newShip.CargoBays {
		return Result{Message: "New ship doesn't have enough cargo bays for your current cargo."}
	}

	gs.Player.Credits -= cost
	gs.Player.Ship = game.Ship{
		TypeID:  shipTypeID,
		Hull:    newShip.Hull,
		Fuel:    newShip.Range,
		Weapons: []int{},
		Shields: []int{},
		Gadgets: []int{},
	}

	return Result{
		Success: true,
		Message: fmt.Sprintf("Purchased %s! Trade-in: %d, Cost: %d.", newShip.Name, tradeIn, cost),
	}
}

func BuyEquipment(gs *game.GameState, equipID int) Result {
	if equipID < 0 || equipID >= len(gs.Data.Equipment) {
		return Result{Message: "Invalid equipment."}
	}

	equip := gs.Data.Equipment[equipID]
	sys := gs.Data.Systems[gs.CurrentSystemID]
	if equip.TechLevel > sys.TechLevel {
		return Result{Message: "Equipment not available at this tech level."}
	}
	if gs.Player.Credits < equip.Price {
		return Result{Message: "Not enough credits."}
	}

	shipDef := gs.Data.Ships[gs.Player.Ship.TypeID]

	switch equip.Category {
	case gamedata.EquipWeapon:
		if len(gs.Player.Ship.Weapons) >= shipDef.WeaponSlots {
			return Result{Message: "No weapon slots available."}
		}
		gs.Player.Ship.Weapons = append(gs.Player.Ship.Weapons, equipID)
	case gamedata.EquipShield:
		if len(gs.Player.Ship.Shields) >= shipDef.ShieldSlots {
			return Result{Message: "No shield slots available."}
		}
		gs.Player.Ship.Shields = append(gs.Player.Ship.Shields, equipID)
	case gamedata.EquipGadget:
		if len(gs.Player.Ship.Gadgets) >= shipDef.GadgetSlots {
			return Result{Message: "No gadget slots available."}
		}
		gs.Player.Ship.Gadgets = append(gs.Player.Ship.Gadgets, equipID)
	}

	gs.Player.Credits -= equip.Price
	return Result{
		Success: true,
		Message: fmt.Sprintf("Installed %s for %d credits.", equip.Name, equip.Price),
	}
}

func SellEquipment(gs *game.GameState, category gamedata.EquipCategory, slotIdx int) Result {
	var slots *[]int
	switch category {
	case gamedata.EquipWeapon:
		slots = &gs.Player.Ship.Weapons
	case gamedata.EquipShield:
		slots = &gs.Player.Ship.Shields
	case gamedata.EquipGadget:
		slots = &gs.Player.Ship.Gadgets
	}

	if slotIdx < 0 || slotIdx >= len(*slots) {
		return Result{Message: "Invalid slot."}
	}

	equipID := (*slots)[slotIdx]
	equip := gs.Data.Equipment[equipID]
	sellPrice := equip.Price * 3 / 4

	*slots = append((*slots)[:slotIdx], (*slots)[slotIdx+1:]...)
	gs.Player.Credits += sellPrice

	return Result{
		Success: true,
		Message: fmt.Sprintf("Sold %s for %d credits.", equip.Name, sellPrice),
	}
}

func MaxHull(gs *game.GameState) int {
	shipDef := gs.Data.Ships[gs.Player.Ship.TypeID]
	maxHull := shipDef.Hull
	if gs.Player.Ship.HullUpgraded {
		maxHull += game.ScarabHullBonus
	}
	return maxHull
}

func RepairCost(gs *game.GameState) int {
	shipDef := gs.Data.Ships[gs.Player.Ship.TypeID]
	damage := MaxHull(gs) - gs.Player.Ship.Hull
	return damage * shipDef.RepairCost
}

func Repair(gs *game.GameState) Result {
	cost := RepairCost(gs)
	if cost == 0 {
		return Result{Message: "Ship is fully repaired."}
	}
	if gs.Player.Credits < cost {
		affordable := gs.Player.Credits / gs.Data.Ships[gs.Player.Ship.TypeID].RepairCost
		if affordable <= 0 {
			return Result{Message: "Not enough credits for any repairs."}
		}
		partialCost := affordable * gs.Data.Ships[gs.Player.Ship.TypeID].RepairCost
		gs.Player.Credits -= partialCost
		gs.Player.Ship.Hull += affordable
		return Result{
			Success: true,
			Message: fmt.Sprintf("Partial repair: +%d hull for %d credits.", affordable, partialCost),
		}
	}

	gs.Player.Credits -= cost
	gs.Player.Ship.Hull = MaxHull(gs)

	return Result{
		Success: true,
		Message: fmt.Sprintf("Fully repaired for %d credits.", cost),
	}
}

const EscapePodPrice = 2000
const InsurancePrice = 1000

func BuyEscapePod(gs *game.GameState) Result {
	if gs.Player.HasEscapePod {
		return Result{Message: "Already have an escape pod."}
	}
	if gs.Player.Credits < EscapePodPrice {
		return Result{Message: "Not enough credits."}
	}
	gs.Player.Credits -= EscapePodPrice
	gs.Player.HasEscapePod = true
	return Result{Success: true, Message: fmt.Sprintf("Escape pod installed for %d credits.", EscapePodPrice)}
}

func BuyInsurance(gs *game.GameState) Result {
	if !gs.Player.HasEscapePod {
		return Result{Message: "Need an escape pod before buying insurance."}
	}
	if gs.Player.HasInsurance {
		return Result{Message: "Already insured."}
	}
	if gs.Player.Credits < InsurancePrice {
		return Result{Message: "Not enough credits."}
	}
	gs.Player.Credits -= InsurancePrice
	gs.Player.HasInsurance = true
	gs.Player.InsuranceDays = 0
	return Result{Success: true, Message: fmt.Sprintf("Insurance purchased for %d credits.", InsurancePrice)}
}

func RefuelCost(gs *game.GameState) int {
	shipDef := gs.Data.Ships[gs.Player.Ship.TypeID]
	maxRange := gs.EffectiveRange()
	needed := maxRange - gs.Player.Ship.Fuel
	return needed * shipDef.FuelCost
}

func Refuel(gs *game.GameState) Result {
	cost := RefuelCost(gs)
	if cost == 0 {
		return Result{Message: "Fuel tank is full."}
	}
	if gs.Player.Credits < cost {
		return Result{Message: "Not enough credits to refuel."}
	}

	gs.Player.Credits -= cost
	gs.Player.Ship.Fuel = gs.EffectiveRange()

	return Result{
		Success: true,
		Message: fmt.Sprintf("Refueled for %d credits.", cost),
	}
}
