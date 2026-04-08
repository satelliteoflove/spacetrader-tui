package encounter

import (
	"fmt"

	"github.com/the4ofus/spacetrader-tui/internal/economy"
	"github.com/the4ofus/spacetrader-tui/internal/formula"
	"github.com/the4ofus/spacetrader-tui/internal/game"
	"github.com/the4ofus/spacetrader-tui/internal/gamedata"
)

func playerWorth(gs *game.GameState) int {
	return economy.PlayerWorth(gs)
}

func Resolve(gs *game.GameState, enc *Encounter, action Action) Outcome {
	switch enc.Type {
	case EncPolice:
		return resolvePolice(gs, action)
	case EncPirate:
		return resolvePirate(gs, enc, action)
	case EncTrader:
		return resolveTrader(gs, enc, action)
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
			fine := IllegalGoodFine * qty
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
	sys := gs.Data.Systems[gs.CurrentSystemID]
	polData := gamedata.PoliticalSystems[sys.PoliticalSystem]

	if polData.BribeLevel <= 0 {
		return Outcome{
			Message: fmt.Sprintf("Bribery is impossible in %s systems.", polData.Name),
		}
	}

	worth := playerWorth(gs)
	diff := int(gs.Difficulty)
	bribeCost := worth / ((BribeBaseDivisor + BribeDiffFactor*(BribeDiffBase-diff)) * polData.BribeLevel)
	bribeCost = (bribeCost / BribeRounding) * BribeRounding
	if bribeCost < MinBribeCost {
		bribeCost = MinBribeCost
	}
	if bribeCost > MaxBribeCost {
		bribeCost = MaxBribeCost
	}

	if gs.Quests.States[game.QuestWild] == game.QuestActive ||
		gs.Quests.States[game.QuestReactor] == game.QuestActive {
		bribeCost *= 2
	}

	if gs.Player.Credits < bribeCost {
		return Outcome{
			Message: fmt.Sprintf("Bribe costs %d credits -- you can't afford it.", bribeCost),
		}
	}

	gs.Player.Credits -= bribeCost
	return Outcome{
		Message:       fmt.Sprintf("Bribe of %d credits accepted. The police look the other way.", bribeCost),
		CreditsChange: -bribeCost,
	}
}

func policeFlee(gs *game.GameState) Outcome {
	enemy := NewPoliceShip(gs)
	gs.Player.PoliceRecord -= 1
	outcome := handleFlee(gs, enemy, -1)
	if !outcome.Fled {
		gs.Player.PoliceRecord -= 2
		outcome.RecordChange = -3
	}
	return outcome
}

func policeSurrender(gs *game.GameState) Outcome {
	attitude := GetPoliceAttitude(gs)

	if attitude == PoliceAttack {
		return confiscateShip(gs)
	}

	record := gs.Player.PoliceRecord
	worth := playerWorth(gs)

	penaltyFactor := -record
	if penaltyFactor > 80 {
		penaltyFactor = 80
	}
	fine := ((1 + worth*penaltyFactor/100) / 500) * 500
	if fine < 500 {
		fine = 500
	}

	prisonDays := -record
	if prisonDays < MinPrisonDays {
		prisonDays = MinPrisonDays
	}

	allCargo := map[int]int{}
	for _, g := range gs.Data.Goods {
		idx := int(g.ID)
		if gs.Player.Cargo[idx] > 0 {
			allCargo[idx] = gs.Player.Cargo[idx]
			gs.Player.Cargo[idx] = 0
		}
	}

	gs.Player.Credits = max(0, gs.Player.Credits-fine)
	gs.Day += prisonDays
	gs.Player.PoliceRecord = ArrestedRecordReset

	msg := fmt.Sprintf("Arrested! Fined %d credits, %d days in prison. All cargo confiscated. Record reset to Dubious.", fine, prisonDays)

	return Outcome{
		Message:       msg,
		CreditsChange: -fine,
		CargoLost:     allCargo,
		RecordChange:  ArrestedRecordReset - record,
	}
}

func confiscateShip(gs *game.GameState) Outcome {
	fleaID := game.ShipFlea
	fleaDef := gs.Data.Ships[fleaID]

	gs.Player.Ship = game.Ship{
		TypeID:  fleaID,
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
		return pirateFlee(gs, enc)
	case ActionSurrender:
		return pirateSurrender(gs)
	}
	return Outcome{Message: "Invalid action."}
}

func pirateFight(gs *game.GameState, enc *Encounter) Outcome {
	var enemy EnemyShip
	if enc.PirateShip != nil {
		enemy = *enc.PirateShip
	} else {
		enemy = NewPirateShip(gs)
	}
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

	lost := min(gs.Player.Credits, PirateLossMin+gs.Rand.Intn(PirateLossRange))
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

func pirateFlee(gs *game.GameState, enc *Encounter) Outcome {
	var enemy EnemyShip
	if enc.PirateShip != nil {
		enemy = *enc.PirateShip
	} else {
		enemy = NewPirateShip(gs)
	}
	return handleFlee(gs, enemy, 0)
}

func handleFlee(gs *game.GameState, enemy EnemyShip, recordChange int) Outcome {
	playerPilot := game.EffectivePlayerSkill(gs, formula.SkillPilot)

	if FleeAttempt(gs.Rand, playerPilot, enemy.PilotSkill, gs.Difficulty) {
		return Outcome{
			Message:      fmt.Sprintf("Escaped the %s!", enemy.Name),
			RecordChange: recordChange,
			Fled:         true,
		}
	}

	round := FleeDamage(gs.Rand, enemy, gs)
	combatLog := FormatCombatLog([]CombatRound{round})

	if destroyed, destroyMsg := checkShipDestroyed(gs); destroyed {
		return Outcome{
			Message:      fmt.Sprintf("Failed to flee! %s", destroyMsg),
			CombatLog:    combatLog,
			HullDamage:   round.HullDamage,
			RecordChange: recordChange,
		}
	}

	return Outcome{
		Message:      fmt.Sprintf("Failed to flee! Took %d hull damage.", round.HullDamage),
		CombatLog:    combatLog,
		HullDamage:   round.HullDamage,
		RecordChange: recordChange,
	}
}

func pirateSurrender(gs *game.GameState) Outcome {
	lost := min(gs.Player.Credits, PirateSurrenderMin+gs.Rand.Intn(PirateSurrenderRange))
	gs.Player.Credits -= lost

	cargoLost := map[int]int{}
	for i := range gs.Player.Cargo {
		if gs.Player.Cargo[i] > 0 && gs.Rand.Intn(CargoLossDenom) < CargoLossChance {
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

func resolveTrader(gs *game.GameState, enc *Encounter, action Action) Outcome {
	switch action {
	case ActionTrade:
		return traderTrade(gs, enc)
	case ActionDecline:
		return Outcome{Message: "Declined to trade."}
	}
	return Outcome{Message: "Invalid action."}
}

func traderTrade(gs *game.GameState, enc *Encounter) Outcome {
	goodIdx := enc.TraderGoodIdx
	price := enc.TraderPrice
	good := gs.Data.Goods[goodIdx]

	dp := &game.GameDataProvider{Data: gs.Data}
	if gs.Player.Credits < price || gs.Player.FreeCargo(dp) < 1 {
		return Outcome{
			Message: fmt.Sprintf("Can't afford %s for %d cr, or no cargo space.", good.Name, price),
		}
	}

	gs.Player.Credits -= price
	gs.Player.Cargo[goodIdx]++

	return Outcome{
		Message:       fmt.Sprintf("Bought 1 %s for %d cr.", good.Name, price),
		CreditsChange: -price,
	}
}

func resolveFamousCaptain(gs *game.GameState, action Action) Outcome {
	if action == ActionDecline {
		return Outcome{Message: "You wave and continue on your way."}
	}

	captainName := ""
	for _, line := range []string{"Captain Ahab", "Captain Conrad", "Captain Huie"} {
		if len(gs.Player.Ship.Shields) > 0 || len(gs.Player.Ship.Weapons) > 0 {
			captainName = line
			break
		}
	}

	isCriminal := gs.Player.PoliceRecord < -30

	if isCriminal {
		return Outcome{Message: "The captain scans your record and moves on without a word."}
	}

	hasReflectiveShield := false
	hasMilitaryLaser := false
	for _, sID := range gs.Player.Ship.Shields {
		if gs.Data.Equipment[sID].Name == "Reflective Shield" || gs.Data.Equipment[sID].Name == "Lightning Shield" {
			hasReflectiveShield = true
		}
	}
	for _, wID := range gs.Player.Ship.Weapons {
		if gs.Data.Equipment[wID].Name == "Military Laser" || gs.Data.Equipment[wID].Name == "Morgan's Laser" {
			hasMilitaryLaser = true
		}
	}

	if captainName == "" {
		captainName = "Captain Ahab"
	}

	switch {
	case captainName == "Captain Ahab" || (!hasMilitaryLaser && hasReflectiveShield):
		if !hasReflectiveShield {
			return Outcome{Message: "Captain Ahab admires your ship but you lack the equipment to learn from him."}
		}
		if gs.Player.Skills[formula.SkillPilot] < formula.SkillMax {
			gs.Player.Skills[formula.SkillPilot]++
			return Outcome{Message: "Captain Ahab shares piloting techniques! Pilot skill increased."}
		}
		return Outcome{Message: "Captain Ahab is impressed by your piloting mastery."}

	case captainName == "Captain Conrad" || (hasMilitaryLaser && gs.Player.Skills[formula.SkillEngineer] < formula.SkillMax):
		if !hasMilitaryLaser {
			return Outcome{Message: "Captain Conrad eyes your weapons but you lack the firepower to impress him."}
		}
		if gs.Player.Skills[formula.SkillEngineer] < formula.SkillMax {
			gs.Player.Skills[formula.SkillEngineer]++
			return Outcome{Message: "Captain Conrad shares engineering secrets! Engineer skill increased."}
		}
		return Outcome{Message: "Captain Conrad salutes your engineering expertise."}

	default:
		if !hasMilitaryLaser {
			return Outcome{Message: "Captain Huie looks at your weapons and shakes his head."}
		}
		if gs.Player.Skills[formula.SkillTrader] < formula.SkillMax {
			gs.Player.Skills[formula.SkillTrader]++
			return Outcome{Message: "Captain Huie shares trade secrets! Trader skill increased."}
		}
		return Outcome{Message: "Captain Huie acknowledges your trading prowess."}
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
	narcIdx := int(gamedata.GoodNarcotics)
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
	skill := gs.Rand.Intn(formula.NumSkills)
	gs.Player.Skills[skill]++
	if gs.Player.Skills[skill] > formula.SkillMax {
		gs.Player.Skills[skill] = formula.SkillMax
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

		gnatID := game.ShipGnat
		gnatDef := gs.Data.Ships[gnatID]
		gs.Player.Ship = game.Ship{
			TypeID:  gnatID,
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
