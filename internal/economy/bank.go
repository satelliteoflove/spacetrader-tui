package economy

import (
	"fmt"

	"github.com/the4ofus/spacetrader-tui/internal/formula"
	"github.com/the4ofus/spacetrader-tui/internal/game"
)

type BankResult struct {
	Success bool
	Message string
}

func TakeLoan(gs *game.GameState, amount int) BankResult {
	if amount <= 0 {
		return BankResult{Message: "Invalid amount."}
	}

	maxLoan := formula.MaxLoan
	available := maxLoan - gs.Player.LoanBalance
	if available <= 0 {
		return BankResult{Message: "You already have the maximum loan."}
	}
	if amount > available {
		amount = available
	}

	gs.Player.LoanBalance += amount
	gs.Player.Credits += amount

	return BankResult{
		Success: true,
		Message: fmt.Sprintf("Borrowed %d credits. Total debt: %d.", amount, gs.Player.LoanBalance),
	}
}

func RepayLoan(gs *game.GameState, amount int) BankResult {
	if amount <= 0 {
		return BankResult{Message: "Invalid amount."}
	}
	if gs.Player.LoanBalance <= 0 {
		return BankResult{Message: "No outstanding loan."}
	}
	if amount > gs.Player.Credits {
		amount = gs.Player.Credits
	}
	if amount > gs.Player.LoanBalance {
		amount = gs.Player.LoanBalance
	}

	gs.Player.Credits -= amount
	gs.Player.LoanBalance -= amount

	if gs.Player.LoanBalance == 0 {
		return BankResult{Success: true, Message: fmt.Sprintf("Repaid %d credits. Loan fully paid!", amount)}
	}
	return BankResult{
		Success: true,
		Message: fmt.Sprintf("Repaid %d credits. Remaining debt: %d.", amount, gs.Player.LoanBalance),
	}
}

func ApplyInterest(gs *game.GameState) int {
	if gs.Player.LoanBalance <= 0 {
		return 0
	}
	interest := gs.Player.LoanBalance / 10
	if interest < 1 {
		interest = 1
	}
	gs.Player.LoanBalance += interest
	return interest
}
