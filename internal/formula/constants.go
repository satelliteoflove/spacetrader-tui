package formula

import "github.com/the4ofus/spacetrader-tui/internal/gamedata"

const (
	MoonPrice       = 500_000
	StartingCredits = 1000
	InterestRate    = 0.10
	StartingShip    = gamedata.ShipGnat

	SkillMin = 1
	SkillMax = 10

	NumSkills = 4

	SkillPilot    = 0
	SkillFighter  = 1
	SkillTrader   = 2
	SkillEngineer = 3
)

var SkillNames = [NumSkills]string{"Pilot", "Fighter", "Trader", "Engineer"}

func SkillPointsForDifficulty(d gamedata.Difficulty) int {
	switch d {
	case gamedata.DiffBeginner:
		return 25
	case gamedata.DiffEasy:
		return 20
	case gamedata.DiffNormal:
		return 20
	case gamedata.DiffHard:
		return 18
	case gamedata.DiffImpossible:
		return 15
	}
	return 20
}
