package encounter

import (
	"github.com/the4ofus/spacetrader-tui/internal/formula"
	"github.com/the4ofus/spacetrader-tui/internal/game"
	"github.com/the4ofus/spacetrader-tui/internal/gamedata"
)

const ClicksPerWarp = 21

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

	threshold := 44 - 2*int(gs.Difficulty)
	if threshold < 1 {
		threshold = 1
	}

	roll := gs.Rand.Intn(threshold)

	if gs.Data.Ships[gs.Player.Ship.TypeID].Name == "Flea" {
		roll *= 2
	}

	if gs.Quests.States[game.QuestAlienArtifact] == game.QuestActive {
		if gs.Rand.Intn(20) < 4 {
			enc := newPirateWithThreat(gs)
			enc.Message = "Alien Mantis ships attack! They want the artifact!"
			return enc
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
		return NewTraderEncounter()
	}

	if gs.Rand.Intn(1000) < 1 {
		return newRareEncounter(gs, 0)
	}
	if gs.Rand.Intn(1000) < 1 {
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
	enc.EnemyPower = piratePowerForDifficulty(gs)
	enc.ThreatNote = assessThreat(gs, enc.EnemyPower)
	return enc
}

func playerCombatPower(gs *game.GameState) int {
	fighterSkill := game.EffectivePlayerSkill(gs, formula.SkillFighter)
	weaponPower := 0
	for _, wID := range gs.Player.Ship.Weapons {
		weaponPower += gs.Data.Equipment[wID].Power
	}
	return fighterSkill*2 + weaponPower
}

func assessThreat(gs *game.GameState, enemyPower int) string {
	player := playerCombatPower(gs)
	if player == 0 && enemyPower == 0 {
		return "Both sides are unarmed."
	}
	ratio := float64(enemyPower) / float64(max(player, 1))
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
