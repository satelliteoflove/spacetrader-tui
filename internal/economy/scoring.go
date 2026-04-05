package economy

import (
	"github.com/the4ofus/spacetrader-tui/internal/game"
	"github.com/the4ofus/spacetrader-tui/internal/gamedata"
)

type ScoreBreakdown struct {
	WorthPoints     int
	DaysPenalty     int
	DiffMultiplier  int
	DiffPercent     int
	FinalScore      int
}

func CalculateScore(gs *game.GameState) ScoreBreakdown {
	netWorth := gs.Player.Credits - gs.Player.LoanBalance
	shipDef := gs.Data.Ships[gs.Player.Ship.TypeID]
	netWorth += shipDef.Price * 3 / 4
	for _, w := range gs.Player.Ship.Weapons {
		netWorth += gs.Data.Equipment[w].Price * 3 / 4
	}
	for _, s := range gs.Player.Ship.Shields {
		netWorth += gs.Data.Equipment[s].Price * 3 / 4
	}
	for _, g := range gs.Player.Ship.Gadgets {
		netWorth += gs.Data.Equipment[g].Price * 3 / 4
	}

	worthPoints := netWorth / 50

	daysPenalty := gs.Day / 2

	diffPercent := 100
	switch gs.Difficulty {
	case gamedata.DiffBeginner:
		diffPercent = 50
	case gamedata.DiffEasy:
		diffPercent = 75
	case gamedata.DiffNormal:
		diffPercent = 100
	case gamedata.DiffHard:
		diffPercent = 130
	case gamedata.DiffImpossible:
		diffPercent = 160
	}

	raw := worthPoints - daysPenalty
	if raw < 0 {
		raw = 0
	}
	finalScore := raw * diffPercent / 100

	return ScoreBreakdown{
		WorthPoints:    worthPoints,
		DaysPenalty:    daysPenalty,
		DiffMultiplier: diffPercent,
		DiffPercent:    diffPercent,
		FinalScore:     finalScore,
	}
}
