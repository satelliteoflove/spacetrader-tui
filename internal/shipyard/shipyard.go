package shipyard

import (
	"fmt"
	"sort"

	"github.com/the4ofus/spacetrader-tui/internal/game"
	"github.com/the4ofus/spacetrader-tui/internal/gamedata"
)

type Result struct {
	Success bool
	Message string
}

type EquipSummary struct {
	Kept      []int
	Sold      []int
	SoldValue int
}

type ShipPurchasePreview struct {
	NewShip     gamedata.ShipDef
	HullTradeIn int
	Weapons     EquipSummary
	Shields     EquipSummary
	Gadgets     EquipSummary
	CrewMustCut int
	NetCost     int
	Error       string
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

func ShipHullTradeIn(gs *game.GameState) int {
	shipDef := gs.Data.Ships[gs.Player.Ship.TypeID]
	mult := 3
	if gs.Quests.TribbleQty > 0 {
		mult = 1
	}
	return shipDef.Price * mult / 4
}

func splitEquipment(equipped []int, maxSlots int, data *gamedata.GameData) (EquipSummary, string) {
	if len(equipped) == 0 {
		return EquipSummary{}, ""
	}

	questCount := 0
	for _, eid := range equipped {
		if data.Equipment[eid].QuestReward {
			questCount++
		}
	}
	if questCount > maxSlots {
		return EquipSummary{}, "New ship can't hold your quest equipment."
	}

	if len(equipped) <= maxSlots {
		kept := make([]int, len(equipped))
		copy(kept, equipped)
		return EquipSummary{Kept: kept}, ""
	}

	var quest, nonQuest []int
	for _, eid := range equipped {
		if data.Equipment[eid].QuestReward {
			quest = append(quest, eid)
		} else {
			nonQuest = append(nonQuest, eid)
		}
	}

	sort.Slice(nonQuest, func(a, b int) bool {
		return data.Equipment[nonQuest[a]].Price > data.Equipment[nonQuest[b]].Price
	})

	remaining := maxSlots - len(quest)
	kept := make([]int, 0, maxSlots)
	kept = append(kept, quest...)

	var sold []int
	var soldValue int

	for i, eid := range nonQuest {
		if i < remaining {
			kept = append(kept, eid)
		} else {
			sold = append(sold, eid)
			soldValue += data.Equipment[eid].Price * 3 / 4
		}
	}

	return EquipSummary{Kept: kept, Sold: sold, SoldValue: soldValue}, ""
}

func PreviewShipPurchase(gs *game.GameState, shipTypeID int) ShipPurchasePreview {
	if shipTypeID < 0 || shipTypeID >= len(gs.Data.Ships) {
		return ShipPurchasePreview{Error: "Invalid ship type."}
	}
	if game.ReactorOnBoard(gs) {
		return ShipPurchasePreview{Error: "Can't trade ships while carrying the Ion Reactor. Deliver it first."}
	}

	newShip := gs.Data.Ships[shipTypeID]
	sys := gs.Data.Systems[gs.CurrentSystemID]
	if newShip.MinTech > sys.TechLevel {
		return ShipPurchasePreview{Error: "Ship not available at this tech level."}
	}

	cargoCount := gs.Player.TotalCargo()
	if cargoCount > newShip.CargoBays {
		return ShipPurchasePreview{Error: "New ship doesn't have enough cargo bays for your current cargo."}
	}

	hullTradeIn := ShipHullTradeIn(gs)

	weapons, errMsg := splitEquipment(gs.Player.Ship.Weapons, newShip.WeaponSlots, gs.Data)
	if errMsg != "" {
		return ShipPurchasePreview{Error: errMsg}
	}
	shields, errMsg := splitEquipment(gs.Player.Ship.Shields, newShip.ShieldSlots, gs.Data)
	if errMsg != "" {
		return ShipPurchasePreview{Error: errMsg}
	}
	gadgets, errMsg := splitEquipment(gs.Player.Ship.Gadgets, newShip.GadgetSlots, gs.Data)
	if errMsg != "" {
		return ShipPurchasePreview{Error: errMsg}
	}

	newMaxCrew := newShip.CrewQuarters - 1
	questCrewCount := 0
	nonQuestCount := 0
	for _, m := range gs.Player.Crew {
		if m.IsQuest {
			questCrewCount++
		} else {
			nonQuestCount++
		}
	}
	if questCrewCount > newMaxCrew {
		return ShipPurchasePreview{Error: "New ship has no quarters for your passenger."}
	}
	crewMustCut := 0
	availableForNonQuest := newMaxCrew - questCrewCount
	if nonQuestCount > availableForNonQuest {
		crewMustCut = nonQuestCount - availableForNonQuest
	}

	equipSoldValue := weapons.SoldValue + shields.SoldValue + gadgets.SoldValue
	netCost := newShip.Price - hullTradeIn - equipSoldValue

	if netCost > gs.Player.Credits {
		return ShipPurchasePreview{Error: fmt.Sprintf("Not enough credits. Need %d more.", netCost-gs.Player.Credits)}
	}

	return ShipPurchasePreview{
		NewShip:     newShip,
		HullTradeIn: hullTradeIn,
		Weapons:     weapons,
		Shields:     shields,
		Gadgets:     gadgets,
		CrewMustCut: crewMustCut,
		NetCost:     netCost,
	}
}

func BuyShip(gs *game.GameState, shipTypeID int, dismissCrew []int) Result {
	preview := PreviewShipPurchase(gs, shipTypeID)
	if preview.Error != "" {
		return Result{Message: preview.Error}
	}

	if len(dismissCrew) != preview.CrewMustCut {
		return Result{Message: fmt.Sprintf("Must dismiss exactly %d crew members.", preview.CrewMustCut)}
	}
	for _, idx := range dismissCrew {
		if idx < 0 || idx >= len(gs.Player.Crew) {
			return Result{Message: "Invalid crew member."}
		}
		if gs.Player.Crew[idx].IsQuest {
			return Result{Message: "Cannot dismiss quest crew."}
		}
	}

	gs.Player.Credits -= preview.NetCost

	sort.Sort(sort.Reverse(sort.IntSlice(dismissCrew)))
	for _, idx := range dismissCrew {
		game.FireMercenary(gs, idx)
	}

	weapons := preview.Weapons.Kept
	if weapons == nil {
		weapons = []int{}
	}
	shields := preview.Shields.Kept
	if shields == nil {
		shields = []int{}
	}
	gadgets := preview.Gadgets.Kept
	if gadgets == nil {
		gadgets = []int{}
	}

	gs.Player.Ship = game.Ship{
		TypeID:  shipTypeID,
		Hull:    preview.NewShip.Hull,
		Fuel:    preview.NewShip.Range,
		Weapons: weapons,
		Shields: shields,
		Gadgets: gadgets,
	}

	return Result{
		Success: true,
		Message: fmt.Sprintf("Purchased %s!", preview.NewShip.Name),
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
	gs.Player.InsuranceDays = 0
	cost := game.InsuranceDailyPremium(gs)
	if gs.Player.Credits < cost {
		return Result{Message: "Not enough credits."}
	}
	gs.Player.Credits -= cost
	gs.Player.HasInsurance = true
	return Result{Success: true, Message: fmt.Sprintf("Insurance purchased for %d credits.", cost)}
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
