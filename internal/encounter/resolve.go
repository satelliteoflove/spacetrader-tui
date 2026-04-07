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
		return resolvePirate(gs, enc, action)
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
	case ActionSurrender:
		return policeSurrender(gs)
	case ActionFight:
		return policeFight(gs)
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

	traderSkill := game.EffectivePlayerSkill(gs, formula.SkillTrader)
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
	playerPilot := game.EffectivePlayerSkill(gs, formula.SkillPilot)
	enemy := NewPoliceShip(gs)

	if FleeAttempt(gs.Rand, playerPilot, enemy.PilotSkill, gs.Difficulty) {
		gs.Player.PoliceRecord -= 1
		return Outcome{
			Message:      "Escaped the police!",
			RecordChange: -1,
			Fled:         true,
		}
	}

	gs.Player.PoliceRecord -= 3
	round := FleeDamage(gs.Rand, enemy, gs)
	combatLog := FormatCombatLog([]CombatRound{round})

	if destroyed, destroyMsg := checkShipDestroyed(gs); destroyed {
		return Outcome{
			Message:      fmt.Sprintf("Failed to flee! %s", destroyMsg),
			CombatLog:    combatLog,
			HullDamage:   round.HullDamage,
			RecordChange: -3,
		}
	}

	return Outcome{
		Message:      fmt.Sprintf("Failed to flee! Took %d hull damage. Record worsened.", round.HullDamage),
		CombatLog:    combatLog,
		HullDamage:   round.HullDamage,
		RecordChange: -3,
	}
}

func policeSurrender(gs *game.GameState) Outcome {
	attitude := GetPoliceAttitude(gs)

	if attitude == PoliceAttack {
		return confiscateShip(gs)
	}

	illegalCargo := map[int]int{}
	allCargo := map[int]int{}
	totalFine := 0

	for _, g := range gs.Data.Goods {
		idx := int(g.ID)
		if gs.Player.Cargo[idx] > 0 {
			if !g.Legal {
				illegalCargo[idx] = gs.Player.Cargo[idx]
				totalFine += 1000 * gs.Player.Cargo[idx]
			}
			allCargo[idx] = gs.Player.Cargo[idx]
		}
	}

	record := gs.Player.PoliceRecord
	if record < -50 {
		for idx, qty := range allCargo {
			gs.Player.Cargo[idx] = 0
			_ = qty
		}
		totalFine += 2000
		gs.Player.Credits = max(0, gs.Player.Credits-totalFine)
		gs.Player.PoliceRecord -= 5
		return Outcome{
			Message:       fmt.Sprintf("All cargo confiscated! Fined %d credits.", totalFine),
			CreditsChange: -totalFine,
			CargoLost:     allCargo,
			RecordChange:  -5,
		}
	}

	for idx := range illegalCargo {
		gs.Player.Cargo[idx] = 0
	}
	if totalFine == 0 {
		totalFine = 500
	}
	gs.Player.Credits = max(0, gs.Player.Credits-totalFine)
	gs.Player.PoliceRecord -= 3
	cargoLost := illegalCargo
	if len(cargoLost) == 0 {
		cargoLost = nil
	}
	return Outcome{
		Message:       fmt.Sprintf("Illegal goods confiscated! Fined %d credits.", totalFine),
		CreditsChange: -totalFine,
		CargoLost:     cargoLost,
		RecordChange:  -3,
	}
}

func confiscateShip(gs *game.GameState) Outcome {
	fleaDef := gs.Data.Ships[0]

	gs.Player.Ship = game.Ship{
		TypeID:  0,
		Hull:    fleaDef.Hull,
		Fuel:    fleaDef.Range,
		Weapons: []int{},
		Shields: []int{},
		Gadgets: []int{},
	}

	for i := range gs.Player.Cargo {
		gs.Player.Cargo[i] = 0
	}
	gs.Player.Crew = nil
	gs.Player.PoliceRecord -= 5

	return Outcome{
		Message:      "Ship confiscated! You have been given a Flea and released.",
		RecordChange: -5,
	}
}

func policeFight(gs *game.GameState) Outcome {
	gs.Player.PoliceRecord -= 5
	enemy := NewPoliceShip(gs)
	result := RunCombat(gs, enemy, 10)
	combatLog := FormatCombatLog(result.Rounds)

	startHull := gs.Data.Ships[gs.Player.Ship.TypeID].Hull
	damage := startHull - gs.Player.Ship.Hull
	if damage < 0 {
		damage = 0
	}

	if result.PlayerWon {
		gs.Player.Reputation++
		gs.Player.Credits += result.Bounty
		return Outcome{
			Message:       fmt.Sprintf("You defeated the police! Bounty: %d cr. Record worsened severely.", result.Bounty),
			CombatLog:     combatLog,
			CreditsChange: result.Bounty,
			HullDamage:    damage,
			RecordChange:  -5,
			RepChange:     1,
		}
	}

	if destroyed, destroyMsg := checkShipDestroyed(gs); destroyed {
		return Outcome{
			Message:      fmt.Sprintf("Defeated by police! %s", destroyMsg),
			CombatLog:    combatLog,
			HullDamage:   damage,
			RecordChange: -5,
		}
	}

	return Outcome{
		Message:      fmt.Sprintf("Defeated by police! Took %d hull damage.", damage),
		CombatLog:    combatLog,
		HullDamage:   damage,
		RecordChange: -5,
	}
}

func resolvePirate(gs *game.GameState, enc *Encounter, action Action) Outcome {
	switch action {
	case ActionFight:
		return pirateFight(gs, enc)
	case ActionFlee:
		return pirateFlee(gs)
	case ActionSurrender:
		return pirateSurrender(gs)
	}
	return Outcome{Message: "Invalid action."}
}

func pirateFight(gs *game.GameState, enc *Encounter) Outcome {
	enemy := NewPirateShip(gs)
	result := RunCombat(gs, enemy, 10)
	combatLog := FormatCombatLog(result.Rounds)

	startHull := gs.Data.Ships[gs.Player.Ship.TypeID].Hull
	damage := startHull - gs.Player.Ship.Hull
	if damage < 0 {
		damage = 0
	}

	if result.PlayerWon {
		gs.Player.Credits += result.Bounty
		gs.Player.Reputation++
		return Outcome{
			Message:       fmt.Sprintf("Victory! Bounty: %d cr.", result.Bounty),
			CombatLog:     combatLog,
			CreditsChange: result.Bounty,
			HullDamage:    damage,
			CargoGained:   result.Loot,
			RepChange:     1,
		}
	}

	lost := min(gs.Player.Credits, 100+gs.Rand.Intn(400))
	gs.Player.Credits -= lost

	if destroyed, destroyMsg := checkShipDestroyed(gs); destroyed {
		return Outcome{
			Message:       fmt.Sprintf("Defeated! %s", destroyMsg),
			CombatLog:     combatLog,
			CreditsChange: -lost,
			HullDamage:    damage,
		}
	}

	return Outcome{
		Message:       fmt.Sprintf("Defeated! Took %d damage, lost %d cr.", damage, lost),
		CombatLog:     combatLog,
		CreditsChange: -lost,
		HullDamage:    damage,
	}
}

func pirateFlee(gs *game.GameState) Outcome {
	playerPilot := game.EffectivePlayerSkill(gs, formula.SkillPilot)
	enemy := NewPirateShip(gs)

	if FleeAttempt(gs.Rand, playerPilot, enemy.PilotSkill, gs.Difficulty) {
		return Outcome{
			Message: "Escaped the pirates!",
			Fled:    true,
		}
	}

	round := FleeDamage(gs.Rand, enemy, gs)
	combatLog := FormatCombatLog([]CombatRound{round})

	if destroyed, destroyMsg := checkShipDestroyed(gs); destroyed {
		return Outcome{
			Message:    fmt.Sprintf("Failed to flee! %s", destroyMsg),
			CombatLog:  combatLog,
			HullDamage: round.HullDamage,
		}
	}

	return Outcome{
		Message:    fmt.Sprintf("Failed to flee! Took %d hull damage.", round.HullDamage),
		CombatLog:  combatLog,
		HullDamage: round.HullDamage,
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

	traderSkill := game.EffectivePlayerSkill(gs, formula.SkillTrader)

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
	gs.Player.Skills[skill]++
	if gs.Player.Skills[skill] > 10 {
		gs.Player.Skills[skill] = 10
	}

	if gs.Rand.Intn(100) < 20 && gs.Quests.TribbleQty == 0 {
		gs.Quests.TribbleQty = 1
		return Outcome{
			Message: fmt.Sprintf("The tonic improved your %s skill! But... what's this furry thing in the bottle?", formula.SkillNames[skill]),
		}
	}

	return Outcome{
		Message: fmt.Sprintf("The tonic improved your %s skill!", formula.SkillNames[skill]),
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
