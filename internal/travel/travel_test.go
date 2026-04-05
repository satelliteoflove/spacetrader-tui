package travel_test

import (
	"os"
	"testing"

	"github.com/the4ofus/spacetrader-tui/internal/data"
	"github.com/the4ofus/spacetrader-tui/internal/formula"
	"github.com/the4ofus/spacetrader-tui/internal/game"
	"github.com/the4ofus/spacetrader-tui/internal/gamedata"
	"github.com/the4ofus/spacetrader-tui/internal/travel"
)

func newTestGame(t *testing.T) *game.GameState {
	t.Helper()
	gd, err := data.LoadAll(os.DirFS("../../data"))
	if err != nil {
		t.Fatalf("LoadAll: %v", err)
	}
	skills := [formula.NumSkills]int{4, 4, 4, 4}
	return game.NewGameWithSeed(gd, "Test", skills, gamedata.DiffNormal, 42)
}

func TestReachableSystems(t *testing.T) {
	gs := newTestGame(t)

	reachable := travel.ReachableSystems(gs)
	if len(reachable) == 0 {
		t.Fatal("no reachable systems from starting position")
	}

	for i := 1; i < len(reachable); i++ {
		if reachable[i].Distance < reachable[i-1].Distance {
			t.Error("reachable systems not sorted by distance")
			break
		}
	}

	for _, r := range reachable {
		if r.Index == gs.CurrentSystemID {
			t.Error("current system should not be in reachable list")
		}
	}
}

func TestTravelDeductsFuel(t *testing.T) {
	gs := newTestGame(t)

	reachable := travel.ReachableSystems(gs)
	if len(reachable) == 0 {
		t.Skip("no reachable systems")
	}

	startFuel := gs.Player.Ship.Fuel
	dest := reachable[0]

	result := travel.ExecuteTravel(gs, dest.Index)
	if !result.Success {
		t.Fatalf("travel failed: %s", result.Message)
	}
	if gs.Player.Ship.Fuel >= startFuel {
		t.Error("fuel should have decreased")
	}
	if result.FuelUsed <= 0 {
		t.Error("fuel used should be positive")
	}
}

func TestTravelAdvancesDay(t *testing.T) {
	gs := newTestGame(t)

	reachable := travel.ReachableSystems(gs)
	if len(reachable) == 0 {
		t.Skip("no reachable systems")
	}

	startDay := gs.Day
	travel.ExecuteTravel(gs, reachable[0].Index)

	if gs.Day != startDay+1 {
		t.Errorf("day: got %d, want %d", gs.Day, startDay+1)
	}
}

func TestTravelMarksVisited(t *testing.T) {
	gs := newTestGame(t)

	reachable := travel.ReachableSystems(gs)
	if len(reachable) == 0 {
		t.Skip("no reachable systems")
	}

	dest := reachable[0]
	travel.ExecuteTravel(gs, dest.Index)

	if !gs.Systems[dest.Index].Visited {
		t.Error("destination not marked as visited")
	}
}

func TestTravelInsufficientFuel(t *testing.T) {
	gs := newTestGame(t)
	gs.Player.Ship.Fuel = 0

	result := travel.ExecuteTravel(gs, 0)
	if result.Success {
		t.Error("should fail with 0 fuel")
	}
}

func TestTravelLoanInterest(t *testing.T) {
	gs := newTestGame(t)
	gs.Player.LoanBalance = 1000

	reachable := travel.ReachableSystems(gs)
	if len(reachable) == 0 {
		t.Skip("no reachable systems")
	}

	travel.ExecuteTravel(gs, reachable[0].Index)

	if gs.Player.LoanBalance != 1100 {
		t.Errorf("loan balance: got %d, want 1100 (1000 + 10%% interest)", gs.Player.LoanBalance)
	}
}

func TestTravelCrewWages(t *testing.T) {
	gs := newTestGame(t)
	gs.Player.Crew = []game.Mercenary{
		{Name: "Test", Skills: [4]int{3, 3, 3, 3}, Wage: 50},
	}

	reachable := travel.ReachableSystems(gs)
	if len(reachable) == 0 {
		t.Skip("no reachable systems")
	}

	startCredits := gs.Player.Credits
	travel.ExecuteTravel(gs, reachable[0].Index)

	if gs.Player.Credits >= startCredits {
		t.Error("credits should decrease from crew wages")
	}
}

func TestTravelCrewFiredWhenBroke(t *testing.T) {
	gs := newTestGame(t)
	gs.Player.Credits = 10
	gs.Player.Crew = []game.Mercenary{
		{Name: "Test", Skills: [4]int{3, 3, 3, 3}, Wage: 50},
	}

	reachable := travel.ReachableSystems(gs)
	if len(reachable) == 0 {
		t.Skip("no reachable systems")
	}

	travel.ExecuteTravel(gs, reachable[0].Index)

	if gs.Player.Credits != 0 {
		t.Errorf("credits should be 0 when broke, got %d", gs.Player.Credits)
	}
	if len(gs.Player.Crew) != 0 {
		t.Error("crew should be dismissed when can't pay wages")
	}
}
