package encounter

import (
	"github.com/the4ofus/spacetrader-tui/internal/game"
	"github.com/the4ofus/spacetrader-tui/internal/gamedata"
)

func Generate(gs *game.GameState) *Encounter {
	sys := gs.Data.Systems[gs.CurrentSystemID]

	policeChance := policeBaseChance(sys.PoliticalSystem)
	pirateChance := pirateBaseChance(sys.PoliticalSystem)
	traderChance := 10

	hasIllegal := false
	for _, g := range gs.Data.Goods {
		if !g.Legal && gs.Player.Cargo[int(g.ID)] > 0 {
			hasIllegal = true
			break
		}
	}
	if hasIllegal {
		policeChance += 20
	}

	if gs.Player.PoliceRecord < -10 {
		policeChance += 10
	}

	if gs.Quests.States[game.QuestWild] == game.QuestActive {
		policeChance += 15
	}

	if gs.Quests.States[game.QuestAlienArtifact] == game.QuestActive && gs.Rand.Intn(100) < 25 {
		return &Encounter{
			Type:    EncPirate,
			Actions: []Action{ActionFight, ActionFlee, ActionSurrender},
			Message: "Alien Mantis ships attack! They want the artifact!",
		}
	}

	roll := gs.Rand.Intn(100)

	if roll < policeChance {
		return NewPoliceEncounter()
	}
	roll -= policeChance

	if roll < pirateChance {
		return NewPirateEncounter()
	}
	roll -= pirateChance

	if roll < traderChance {
		return NewTraderEncounter()
	}
	roll -= traderChance

	if roll < 2 {
		return newRareEncounter(gs)
	}

	return nil
}

func newRareEncounter(gs *game.GameState) *Encounter {
	roll := gs.Rand.Intn(3)
	switch roll {
	case 0:
		captains := []string{"Captain Ahab", "Captain Conrad", "Captain Huie"}
		name := captains[gs.Rand.Intn(len(captains))]
		return &Encounter{
			Type:    EncFamousCaptain,
			Actions: []Action{ActionTrade, ActionDecline},
			Message: name + " hails you. The famous captain offers supplies.",
		}
	case 1:
		return &Encounter{
			Type:    EncMarieCeleste,
			Actions: []Action{ActionTrade, ActionDecline},
			Message: "You find an abandoned ship, the Marie Celeste. Its cargo hold contains narcotics.",
		}
	case 2:
		return &Encounter{
			Type:    EncBottle,
			Actions: []Action{ActionTrade, ActionDecline},
			Message: "You find a floating bottle containing Captain Marmoset's Skill Tonic!",
		}
	}
	return nil
}

func policeBaseChance(pol gamedata.PoliticalSystem) int {
	switch pol {
	case gamedata.PolAnarchy, gamedata.PolFeudal:
		return 0
	case gamedata.PolMilitary, gamedata.PolFascist:
		return 30
	case gamedata.PolCorporate, gamedata.PolCybernetic, gamedata.PolTechnocracy:
		return 25
	case gamedata.PolDictatorship:
		return 20
	case gamedata.PolDemocracy, gamedata.PolCapitalist, gamedata.PolConfederacy:
		return 15
	case gamedata.PolMonarchy, gamedata.PolTheocracy:
		return 15
	case gamedata.PolSocialist, gamedata.PolCommunist:
		return 10
	case gamedata.PolPacifist, gamedata.PolSatori:
		return 5
	}
	return 15
}

func pirateBaseChance(pol gamedata.PoliticalSystem) int {
	switch pol {
	case gamedata.PolAnarchy, gamedata.PolFeudal:
		return 30
	case gamedata.PolSocialist, gamedata.PolCommunist:
		return 20
	case gamedata.PolDemocracy, gamedata.PolCapitalist, gamedata.PolConfederacy:
		return 15
	case gamedata.PolMonarchy, gamedata.PolTheocracy:
		return 15
	case gamedata.PolDictatorship:
		return 10
	case gamedata.PolCorporate, gamedata.PolCybernetic, gamedata.PolTechnocracy:
		return 10
	case gamedata.PolMilitary, gamedata.PolFascist:
		return 5
	case gamedata.PolPacifist, gamedata.PolSatori:
		return 5
	}
	return 15
}
