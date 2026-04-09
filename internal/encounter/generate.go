package encounter

import (
	"fmt"

	"github.com/the4ofus/spacetrader-tui/internal/formula"
	"github.com/the4ofus/spacetrader-tui/internal/game"
	"github.com/the4ofus/spacetrader-tui/internal/gamedata"
)

type PoliceAttitude int

const (
	PoliceInspect   PoliceAttitude = iota
	PoliceSurrender
	PoliceAttack
)

func GetPoliceAttitude(gs *game.GameState) PoliceAttitude {
	record := gs.Player.PoliceRecord
	diff := int(gs.Difficulty)

	surrenderThresholds := [5]int{-999, -70, -30, -10, 0}
	attackThresholds := [5]int{-999, -100, -70, -30, -10}

	if diff > 4 {
		diff = 4
	}

	if record < attackThresholds[diff] {
		return PoliceAttack
	}
	if record < surrenderThresholds[diff] {
		return PoliceSurrender
	}
	return PoliceInspect
}

func GenerateForClick(gs *game.GameState, destIdx int) *Encounter {
	dest := gs.Data.Systems[destIdx]
	polData := gamedata.PoliticalSystems[dest.PoliticalSystem]

	threshold := EncounterBaseThreshold - DifficultyThresholdMod*int(gs.Difficulty)
	if threshold < 1 {
		threshold = 1
	}

	roll := gs.Rand.Intn(threshold)

	if gs.Data.Ships[gs.Player.Ship.TypeID].Name == "Flea" {
		roll *= 2
	}

	if gs.Quests.States[game.QuestAlienArtifact] == game.QuestActive {
		if gs.Rand.Intn(AlienArtifactDenom) < AlienArtifactChance {
			enc := newPirateWithThreat(gs)
			enc.Message = "Alien Mantis ships attack! They want the artifact!"
			return enc
		}
	}

	if gs.Quests.States[game.QuestWild] == game.QuestActive {
		kravat := -1
		for i, sys := range gs.Data.Systems {
			if sys.Name == "Kravat" {
				kravat = i
				break
			}
		}
		if kravat >= 0 && destIdx == kravat {
			policeChance := 100 / max(2, min(4, 5-int(gs.Difficulty)))
			if gs.Rand.Intn(100) < policeChance {
				return newPoliceForAttitude(gs)
			}
		}
	}

	pirateStrength := polData.PirateStrength
	policeStrength := polData.PoliceStrength
	traderStrength := polData.TraderStrength

	if roll < pirateStrength {
		return newPirateWithThreat(gs)
	}

	if roll < pirateStrength+policeStrength {
		return newPoliceForAttitude(gs)
	}

	if roll < pirateStrength+policeStrength+traderStrength {
		return newTraderWithOffer(gs, destIdx)
	}

	if gs.Rand.Intn(RareEncounterOdds) < 1 {
		return newRareEncounter(gs, 0)
	}
	if gs.Rand.Intn(RareEncounterOdds) < 1 {
		return newRareEncounter(gs, 1)
	}

	return nil
}

func Generate(gs *game.GameState) *Encounter {
	return GenerateForClick(gs, gs.CurrentSystemID)
}

func newRareEncounter(gs *game.GameState, variant int) *Encounter {
	switch variant {
	case 0:
		captains := []string{"Captain Ahab", "Captain Conrad", "Captain Huie"}
		name := captains[gs.Rand.Intn(len(captains))]
		return &Encounter{
			Type:    EncFamousCaptain,
			Actions: []Action{ActionTrade, ActionDecline},
			Message: name + " hails you. The famous captain offers to share wisdom.",
		}
	case 1:
		if gs.Rand.Intn(2) == 0 {
			return &Encounter{
				Type:    EncMarieCeleste,
				Actions: []Action{ActionTrade, ActionDecline},
				Message: "You find an abandoned ship, the Marie Celeste. Its cargo hold contains narcotics.",
			}
		}
		return &Encounter{
			Type:    EncBottle,
			Actions: []Action{ActionTrade, ActionDecline},
			Message: "You find a floating bottle containing Captain Marmoset's Skill Tonic!",
		}
	}
	return nil
}

func newPoliceForAttitude(gs *game.GameState) *Encounter {
	attitude := GetPoliceAttitude(gs)
	switch attitude {
	case PoliceAttack:
		return &Encounter{
			Type:    EncPolice,
			Actions: []Action{ActionFight, ActionFlee},
			Message: "Police open fire on sight! Your record precedes you.",
		}
	case PoliceSurrender:
		return &Encounter{
			Type:    EncPolice,
			Actions: []Action{ActionSurrender, ActionFight, ActionFlee},
			Message: "Police demand your immediate surrender!",
		}
	default:
		return NewPoliceEncounter()
	}
}

func newPirateWithThreat(gs *game.GameState) *Encounter {
	enc := NewPirateEncounter()
	ship := NewPirateShip(gs)
	enc.PirateShip = &ship
	enc.ThreatNote = assessThreat(gs, &ship)
	return enc
}

func shipCombatPower(weaponPower int, fighterSkill int, hull int, totalShields int) int {
	return weaponPower*2 + fighterSkill*2 + hull/10 + totalShields/10
}

func assessThreat(gs *game.GameState, enemy *EnemyShip) string {
	playerFighter := game.EffectivePlayerSkill(gs, formula.SkillFighter)
	playerWeapon := 0
	for _, wID := range gs.Player.Ship.Weapons {
		playerWeapon += gs.Data.Equipment[wID].Power
	}
	playerHull := gs.Data.Ships[gs.Player.Ship.TypeID].Hull
	playerShields := 0
	for _, sID := range gs.Player.Ship.Shields {
		playerShields += gs.Data.Equipment[sID].Protection
	}
	player := shipCombatPower(playerWeapon, playerFighter, playerHull, playerShields)

	enemyShields := 0
	for _, s := range enemy.Shields {
		enemyShields += s
	}
	pirate := shipCombatPower(enemy.WeaponPower, enemy.FighterSkill, enemy.Hull, enemyShields)

	if player == 0 && pirate == 0 {
		return "Both sides are unarmed."
	}
	ratio := float64(pirate) / float64(max(player, 1))
	switch {
	case ratio <= 0.5:
		return "Your scanners detect a lightly armed vessel."
	case ratio <= 0.8:
		return "The pirate appears outmatched."
	case ratio <= 1.2:
		return "The pirate appears evenly matched."
	case ratio <= 1.8:
		return "A heavily armed pirate -- dangerous."
	default:
		return "This pirate outguns you significantly."
	}
}

func newTraderWithOffer(gs *game.GameState, destIdx int) *Encounter {
	sysState := &gs.Systems[destIdx]
	var available []int
	for i := 0; i < game.NumGoods; i++ {
		if sysState.Prices[i] > 0 {
			available = append(available, i)
		}
	}
	if len(available) == 0 {
		return NewTraderEncounter()
	}

	goodIdx := available[gs.Rand.Intn(len(available))]
	good := gs.Data.Goods[goodIdx]

	traderSkill := game.EffectivePlayerSkill(gs, formula.SkillTrader)
	discount := 90 + traderSkill
	if discount > 98 {
		discount = 98
	}
	price := sysState.Prices[goodIdx] * discount / 100

	return &Encounter{
		Type:          EncTrader,
		Actions:       []Action{ActionTrade, ActionDecline},
		Message:       fmt.Sprintf("A trader offers 1 %s for %d cr (market: %d cr).", good.Name, price, sysState.Prices[goodIdx]),
		TraderGoodIdx: goodIdx,
		TraderPrice:   price,
	}
}
