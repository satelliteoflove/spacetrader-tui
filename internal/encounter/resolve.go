package encounter

import (
	"fmt"

	"github.com/the4ofus/spacetrader-tui/internal/formula"
	"github.com/the4ofus/spacetrader-tui/internal/game"
)

func Resolve(gs *game.GameState, enc *Encounter, action Action) Outcome {
	switch enc.Type {
	case EncPolice:
		return resolvePolice(gs, action)
	case EncPirate:
		return resolvePirate(gs, action)
	case EncTrader:
		return resolveTrader(gs, action)
	case EncFamousCaptain:
		return resolveFamousCaptain(gs, action)
	case EncMarieCeleste:
		return resolveMarieCeleste(gs, action)
	case EncBottle:
		return resolveBottle(gs, action)
	}
	return Outcome{Message: "Nothing happened."}
}

func resolvePolice(gs *game.GameState, action Action) Outcome {
	switch action {
	case ActionComply:
		return policeComply(gs)
	case ActionBribe:
		return policeBribe(gs)
	case ActionFlee:
		return policeFlee(gs)
	}
	return Outcome{Message: "Invalid action."}
}

func policeComply(gs *game.GameState) Outcome {
	illegalCargo := map[int]int{}
	totalFine := 0

	for _, g := range gs.Data.Goods {
		idx := int(g.ID)
		if !g.Legal && gs.Player.Cargo[idx] > 0 {
			qty := gs.Player.Cargo[idx]
			fine := 500 * qty
			totalFine += fine
			illegalCargo[idx] = qty
			gs.Player.Cargo[idx] = 0
		}
	}

	if totalFine > 0 {
		gs.Player.Credits = max(0, gs.Player.Credits-totalFine)
		gs.Player.PoliceRecord -= 2
		return Outcome{
			Message:       fmt.Sprintf("Illegal goods confiscated! Fined %d credits.", totalFine),
			CreditsChange: -totalFine,
			CargoLost:     illegalCargo,
			RecordChange:  -2,
		}
	}

	gs.Player.PoliceRecord++
	return Outcome{
		Message:      "Police found nothing illegal. Record improved.",
		RecordChange: 1,
	}
}

func policeBribe(gs *game.GameState) Outcome {
	bribeAmount := 500 + gs.Rand.Intn(1000)
	if gs.Player.Credits < bribeAmount {
		gs.Player.PoliceRecord -= 1
		return Outcome{
			Message:      "Can't afford the bribe. Police record worsened.",
			RecordChange: -1,
		}
	}

	crewMercs := make([]formula.Mercenary, len(gs.Player.Crew))
	for i, m := range gs.Player.Crew {
		crewMercs[i] = m
	}
	traderSkill := formula.EffectiveSkill(gs.Player.Skills[formula.SkillTrader], crewMercs, formula.SkillTrader, 0)
	successChance := 20 + traderSkill*5

	gs.Player.Credits -= bribeAmount

	if gs.Rand.Intn(100) < successChance {
		return Outcome{
			Message:       fmt.Sprintf("Bribe of %d credits accepted.", bribeAmount),
			CreditsChange: -bribeAmount,
		}
	}

	gs.Player.PoliceRecord -= 2
	return Outcome{
		Message:       fmt.Sprintf("Bribe of %d credits rejected! Record worsened.", bribeAmount),
		CreditsChange: -bribeAmount,
		RecordChange:  -2,
	}
}

func policeFlee(gs *game.GameState) Outcome {
	crewMercs := make([]formula.Mercenary, len(gs.Player.Crew))
	for i, m := range gs.Player.Crew {
		crewMercs[i] = m
	}
	pilotSkill := formula.EffectiveSkill(gs.Player.Skills[formula.SkillPilot], crewMercs, formula.SkillPilot, 0)
	fleeChance := 30 + pilotSkill*5

	if gs.Rand.Intn(100) < fleeChance {
		gs.Player.PoliceRecord -= 1
		return Outcome{
			Message:      "Escaped the police!",
			RecordChange: -1,
			Fled:         true,
		}
	}

	gs.Player.PoliceRecord -= 3
	return Outcome{
		Message:      "Failed to flee. Police record worsened significantly.",
		RecordChange: -3,
	}
}

func resolvePirate(gs *game.GameState, action Action) Outcome {
	switch action {
	case ActionFight:
		return pirateFight(gs)
	case ActionFlee:
		return pirateFlee(gs)
	case ActionSurrender:
		return pirateSurrender(gs)
	}
	return Outcome{Message: "Invalid action."}
}

func pirateFight(gs *game.GameState) Outcome {
	crewMercs := make([]formula.Mercenary, len(gs.Player.Crew))
	for i, m := range gs.Player.Crew {
		crewMercs[i] = m
	}
	fighterSkill := formula.EffectiveSkill(gs.Player.Skills[formula.SkillFighter], crewMercs, formula.SkillFighter, 0)

	weaponPower := 0
	for _, wID := range gs.Player.Ship.Weapons {
		weaponPower += gs.Data.Equipment[wID].Power
	}

	playerPower := fighterSkill*2 + weaponPower
	piratePower := piratePowerForDifficulty(gs)

	if playerPower >= piratePower {
		loot := 200 + gs.Rand.Intn(1800)
		gs.Player.Credits += loot
		gs.Player.Reputation++
		return Outcome{
			Message:       fmt.Sprintf("Victory! Looted %d credits. (You: %d vs Pirate: %d)", loot, playerPower, piratePower),
			CreditsChange: loot,
			RepChange:     1,
		}
	}

	shieldProtection := 0
	for _, sID := range gs.Player.Ship.Shields {
		shieldProtection += gs.Data.Equipment[sID].Protection
	}

	rawDamage := 10 + gs.Rand.Intn(30)
	damage := rawDamage - shieldProtection/10
	if damage < 0 {
		damage = 0
	}
	gs.Player.Ship.Hull -= damage

	lost := min(gs.Player.Credits, 100+gs.Rand.Intn(400))
	gs.Player.Credits -= lost

	if destroyed, destroyMsg := checkShipDestroyed(gs); destroyed {
		return Outcome{
			Message:       fmt.Sprintf("Defeated! (You: %d vs Pirate: %d) %s", playerPower, piratePower, destroyMsg),
			CreditsChange: -lost,
			HullDamage:    damage,
		}
	}

	return Outcome{
		Message:       fmt.Sprintf("Defeated! Took %d damage, lost %d credits. (You: %d vs Pirate: %d)", damage, lost, playerPower, piratePower),
		CreditsChange: -lost,
		HullDamage:    damage,
	}
}

func pirateFlee(gs *game.GameState) Outcome {
	crewMercs := make([]formula.Mercenary, len(gs.Player.Crew))
	for i, m := range gs.Player.Crew {
		crewMercs[i] = m
	}
	pilotSkill := formula.EffectiveSkill(gs.Player.Skills[formula.SkillPilot], crewMercs, formula.SkillPilot, 0)
	fleeChance := 30 + pilotSkill*5

	if gs.Rand.Intn(100) < fleeChance {
		return Outcome{
			Message: "Escaped the pirates!",
			Fled:    true,
		}
	}

	shieldProtection := 0
	for _, sID := range gs.Player.Ship.Shields {
		shieldProtection += gs.Data.Equipment[sID].Protection
	}

	rawDamage := 5 + gs.Rand.Intn(15)
	damage := rawDamage - shieldProtection/10
	if damage < 0 {
		damage = 0
	}
	gs.Player.Ship.Hull -= damage

	lost := min(gs.Player.Credits, 50+gs.Rand.Intn(250))
	gs.Player.Credits -= lost

	if destroyed, destroyMsg := checkShipDestroyed(gs); destroyed {
		return Outcome{
			Message:       fmt.Sprintf("Failed to flee! %s", destroyMsg),
			CreditsChange: -lost,
			HullDamage:    damage,
		}
	}

	return Outcome{
		Message:       fmt.Sprintf("Failed to flee! Took %d damage, lost %d credits.", damage, lost),
		CreditsChange: -lost,
		HullDamage:    damage,
	}
}

func pirateSurrender(gs *game.GameState) Outcome {
	lost := min(gs.Player.Credits, 200+gs.Rand.Intn(800))
	gs.Player.Credits -= lost

	cargoLost := map[int]int{}
	for i := range gs.Player.Cargo {
		if gs.Player.Cargo[i] > 0 && gs.Rand.Intn(100) < 50 {
			cargoLost[i] = gs.Player.Cargo[i]
			gs.Player.Cargo[i] = 0
		}
	}

	return Outcome{
		Message:       fmt.Sprintf("Surrendered. Lost %d credits and some cargo.", lost),
		CreditsChange: -lost,
		CargoLost:     cargoLost,
	}
}

func resolveTrader(gs *game.GameState, action Action) Outcome {
	switch action {
	case ActionTrade:
		return traderTrade(gs)
	case ActionDecline:
		return Outcome{Message: "Declined to trade."}
	}
	return Outcome{Message: "Invalid action."}
}

func traderTrade(gs *game.GameState) Outcome {
	sysState := &gs.Systems[gs.CurrentSystemID]

	available := []int{}
	for i := 0; i < game.NumGoods; i++ {
		if sysState.Prices[i] > 0 {
			available = append(available, i)
		}
	}
	if len(available) == 0 {
		return Outcome{Message: "Trader has nothing to offer."}
	}

	goodIdx := available[gs.Rand.Intn(len(available))]
	good := gs.Data.Goods[goodIdx]

	crewMercs := make([]formula.Mercenary, len(gs.Player.Crew))
	for i, m := range gs.Player.Crew {
		crewMercs[i] = m
	}
	traderSkill := formula.EffectiveSkill(gs.Player.Skills[formula.SkillTrader], crewMercs, formula.SkillTrader, 0)

	discount := 90 + traderSkill
	if discount > 98 {
		discount = 98
	}
	price := sysState.Prices[goodIdx] * discount / 100

	dp := &game.GameDataProvider{Data: gs.Data}
	if gs.Player.Credits < price || gs.Player.FreeCargo(dp) < 1 {
		return Outcome{
			Message: fmt.Sprintf("Trader offered %s for %d credits, but you can't afford it or have no space.", good.Name, price),
		}
	}

	gs.Player.Credits -= price
	gs.Player.Cargo[goodIdx]++

	return Outcome{
		Message:       fmt.Sprintf("Bought 1 %s from trader for %d credits (market price: %d).", good.Name, price, sysState.Prices[goodIdx]),
		CreditsChange: -price,
	}
}

func resolveFamousCaptain(gs *game.GameState, action Action) Outcome {
	if action == ActionDecline {
		return Outcome{Message: "You wave and continue on your way."}
	}
	repair := 10 + gs.Rand.Intn(20)
	shipDef := gs.Data.Ships[gs.Player.Ship.TypeID]
	gs.Player.Ship.Hull += repair
	if gs.Player.Ship.Hull > shipDef.Hull {
		gs.Player.Ship.Hull = shipDef.Hull
	}
	return Outcome{
		Message: fmt.Sprintf("The captain repairs your hull (+%d) as a gesture of goodwill.", repair),
	}
}

func resolveMarieCeleste(gs *game.GameState, action Action) Outcome {
	if action == ActionDecline {
		return Outcome{Message: "You leave the derelict alone."}
	}
	dp := &game.GameDataProvider{Data: gs.Data}
	if gs.Player.FreeCargo(dp) < 3 {
		return Outcome{Message: "Not enough cargo space to salvage."}
	}
	narcIdx := 8
	gs.Player.Cargo[narcIdx] += 3
	gs.Player.PoliceRecord -= 1
	return Outcome{
		Message:      "Salvaged 3 units of narcotics. Police may ask questions...",
		RecordChange: -1,
	}
}

func resolveBottle(gs *game.GameState, action Action) Outcome {
	if action == ActionDecline {
		return Outcome{Message: "You leave the bottle floating in space."}
	}
	skill := gs.Rand.Intn(4)
	skillNames := []string{"Pilot", "Fighter", "Trader", "Engineer"}
	gs.Player.Skills[skill]++
	if gs.Player.Skills[skill] > 10 {
		gs.Player.Skills[skill] = 10
	}

	if gs.Rand.Intn(100) < 20 && gs.Quests.TribbleQty == 0 {
		gs.Quests.TribbleQty = 1
		return Outcome{
			Message: fmt.Sprintf("The tonic improved your %s skill! But... what's this furry thing in the bottle?", skillNames[skill]),
		}
	}

	return Outcome{
		Message: fmt.Sprintf("The tonic improved your %s skill!", skillNames[skill]),
	}
}

func piratePowerForDifficulty(gs *game.GameState) int {
	base := 10
	spread := 40
	switch gs.Difficulty {
	case 0: // Beginner
		base = 5
		spread = 20
	case 1: // Easy
		base = 8
		spread = 30
	case 2: // Normal
		base = 10
		spread = 40
	case 3: // Hard
		base = 15
		spread = 50
	case 4: // Impossible
		base = 25
		spread = 60
	}
	dayBonus := gs.Day / 20
	return base + gs.Rand.Intn(spread) + dayBonus
}

func checkShipDestroyed(gs *game.GameState) (destroyed bool, message string) {
	if gs.Player.Ship.Hull > 0 {
		return false, ""
	}
	gs.Player.Ship.Hull = 0

	if gs.Player.HasEscapePod {
		gs.Player.HasEscapePod = false
		gs.Player.Crew = nil

		insurancePayout := 0
		if gs.Player.HasInsurance {
			shipDef := gs.Data.Ships[gs.Player.Ship.TypeID]
			insurancePayout = shipDef.Price
			for _, w := range gs.Player.Ship.Weapons {
				insurancePayout += gs.Data.Equipment[w].Price
			}
			for _, s := range gs.Player.Ship.Shields {
				insurancePayout += gs.Data.Equipment[s].Price
			}
			for _, g := range gs.Player.Ship.Gadgets {
				insurancePayout += gs.Data.Equipment[g].Price
			}
			insurancePayout = insurancePayout * 3 / 4
			gs.Player.Credits += insurancePayout
			gs.Player.HasInsurance = false
		}

		gnatDef := gs.Data.Ships[1]
		gs.Player.Ship = game.Ship{
			TypeID:  1,
			Hull:    gnatDef.Hull,
			Fuel:    gnatDef.Range,
			Weapons: []int{},
			Shields: []int{},
			Gadgets: []int{},
		}

		for i := range gs.Player.Cargo {
			gs.Player.Cargo[i] = 0
		}

		msg := "Ship destroyed! Escape pod activated."
		if insurancePayout > 0 {
			msg += fmt.Sprintf(" Insurance paid %d credits.", insurancePayout)
		}
		msg += " You start over with a Gnat."
		return true, msg
	}

	gs.EndStatus = game.StatusDead
	return true, "Ship destroyed!"
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
