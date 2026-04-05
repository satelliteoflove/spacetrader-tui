package economy_test

import (
	"os"
	"testing"

	"github.com/the4ofus/spacetrader-tui/internal/data"
	"github.com/the4ofus/spacetrader-tui/internal/economy"
	"github.com/the4ofus/spacetrader-tui/internal/formula"
	"github.com/the4ofus/spacetrader-tui/internal/game"
	"github.com/the4ofus/spacetrader-tui/internal/gamedata"
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

func TestTakeLoan(t *testing.T) {
	gs := newTestGame(t)
	startCredits := gs.Player.Credits

	result := economy.TakeLoan(gs, 5000)
	if !result.Success {
		t.Fatalf("TakeLoan failed: %s", result.Message)
	}
	if gs.Player.Credits != startCredits+5000 {
		t.Errorf("credits: got %d, want %d", gs.Player.Credits, startCredits+5000)
	}
	if gs.Player.LoanBalance != 5000 {
		t.Errorf("loan: got %d, want 5000", gs.Player.LoanBalance)
	}
}

func TestTakeLoanMax(t *testing.T) {
	gs := newTestGame(t)
	gs.Player.LoanBalance = 25000

	result := economy.TakeLoan(gs, 1000)
	if result.Success {
		t.Error("should not be able to take loan at max")
	}
}

func TestRepayLoan(t *testing.T) {
	gs := newTestGame(t)
	gs.Player.LoanBalance = 5000
	gs.Player.Credits = 10000

	result := economy.RepayLoan(gs, 3000)
	if !result.Success {
		t.Fatalf("RepayLoan failed: %s", result.Message)
	}
	if gs.Player.LoanBalance != 2000 {
		t.Errorf("loan: got %d, want 2000", gs.Player.LoanBalance)
	}
	if gs.Player.Credits != 7000 {
		t.Errorf("credits: got %d, want 7000", gs.Player.Credits)
	}
}

func TestRepayLoanFull(t *testing.T) {
	gs := newTestGame(t)
	gs.Player.LoanBalance = 1000
	gs.Player.Credits = 5000

	result := economy.RepayLoan(gs, 5000)
	if !result.Success {
		t.Fatalf("RepayLoan failed: %s", result.Message)
	}
	if gs.Player.LoanBalance != 0 {
		t.Errorf("loan: got %d, want 0", gs.Player.LoanBalance)
	}
	if gs.Player.Credits != 4000 {
		t.Errorf("credits: got %d, want 4000", gs.Player.Credits)
	}
}

func TestApplyInterest(t *testing.T) {
	gs := newTestGame(t)
	gs.Player.LoanBalance = 1000

	interest := economy.ApplyInterest(gs)
	if interest != 100 {
		t.Errorf("interest: got %d, want 100", interest)
	}
	if gs.Player.LoanBalance != 1100 {
		t.Errorf("loan: got %d, want 1100", gs.Player.LoanBalance)
	}
}

func TestApplyInterestNoLoan(t *testing.T) {
	gs := newTestGame(t)

	interest := economy.ApplyInterest(gs)
	if interest != 0 {
		t.Errorf("interest: got %d, want 0", interest)
	}
}
