package travel

import (
	"github.com/the4ofus/spacetrader-tui/internal/economy"
	"github.com/the4ofus/spacetrader-tui/internal/game"
)

type DailyCost struct {
	Wages     int
	Interest  int
	Premium   int
}

func (d DailyCost) Total() int {
	return d.Wages + d.Interest + d.Premium
}

func NextDayCost(gs *game.GameState) DailyCost {
	var d DailyCost
	for _, m := range gs.Player.Crew {
		d.Wages += m.Wage()
	}
	if gs.Player.LoanBalance > 0 {
		d.Interest = economy.LoanInterest(gs.Player.LoanBalance)
	}
	if gs.Player.HasInsurance {
		d.Premium = game.InsuranceDailyPremium(gs)
	}
	return d
}
