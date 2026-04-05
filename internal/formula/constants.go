package formula

import "github.com/the4ofus/spacetrader-tui/internal/gamedata"

const (
	MoonPrice       = 500_000
	StartingCredits = 1000
	MaxLoan         = 25_000
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

func SkillPointsForDifficulty(d gamedata.Difficulty) int {
	switch d {
	case gamedata.DiffBeginner:
		return 20
	case gamedata.DiffEasy:
		return 18
	case gamedata.DiffNormal:
		return 16
	case gamedata.DiffHard:
		return 14
	case gamedata.DiffImpossible:
		return 12
	}
	return 16
}

func MaxLoanForDifficulty(d gamedata.Difficulty) int {
	switch d {
	case gamedata.DiffBeginner:
		return 25_000
	case gamedata.DiffEasy:
		return 25_000
	case gamedata.DiffNormal:
		return 25_000
	case gamedata.DiffHard:
		return 25_000
	case gamedata.DiffImpossible:
		return 25_000
	}
	return 25_000
}
