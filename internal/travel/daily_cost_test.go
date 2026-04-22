package travel_test

import (
	"testing"

	"github.com/the4ofus/spacetrader-tui/internal/game"
	"github.com/the4ofus/spacetrader-tui/internal/travel"
)

func TestNextDayCostEmpty(t *testing.T) {
	gs := newTestGame(t)
	gs.Player.Crew = nil
	gs.Player.LoanBalance = 0
	gs.Player.HasInsurance = false

	c := travel.NextDayCost(gs)
	if c.Total() != 0 {
		t.Errorf("expected zero cost, got %+v", c)
	}
}

func TestNextDayCostLoanOnly(t *testing.T) {
	gs := newTestGame(t)
	gs.Player.Crew = nil
	gs.Player.LoanBalance = 1000
	gs.Player.HasInsurance = false

	c := travel.NextDayCost(gs)
	if c.Interest != 100 {
		t.Errorf("expected 100 interest on 1000 loan, got %d", c.Interest)
	}
	if c.Wages != 0 || c.Premium != 0 {
		t.Errorf("only interest should be nonzero, got %+v", c)
	}
}

func TestNextDayCostInsuranceOnly(t *testing.T) {
	gs := newTestGame(t)
	gs.Player.Crew = nil
	gs.Player.LoanBalance = 0
	gs.Player.HasInsurance = true

	c := travel.NextDayCost(gs)
	if c.Premium <= 0 {
		t.Errorf("expected nonzero premium, got %d", c.Premium)
	}
	want := game.InsuranceDailyPremium(gs)
	if c.Premium != want {
		t.Errorf("premium mismatch: got %d, want %d", c.Premium, want)
	}
}

func TestNextDayCostCrewWages(t *testing.T) {
	gs := newTestGame(t)
	gs.Player.LoanBalance = 0
	gs.Player.HasInsurance = false
	gs.Player.Crew = []game.Mercenary{
		{Skills: [4]int{5, 5, 5, 5}},
		{Skills: [4]int{8, 2, 2, 2}},
	}

	c := travel.NextDayCost(gs)
	expectedWages := gs.Player.Crew[0].Wage() + gs.Player.Crew[1].Wage()
	if c.Wages != expectedWages {
		t.Errorf("wages: got %d, want %d", c.Wages, expectedWages)
	}
}
