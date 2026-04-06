package market

import (
	"fmt"

	"github.com/the4ofus/spacetrader-tui/internal/formula"
	"github.com/the4ofus/spacetrader-tui/internal/game"
)

type TransactionResult struct {
	Success    bool
	Message    string
	TotalPrice int
}

func effectiveTraderSkill(gs *game.GameState) int {
	return formula.EffectiveSkill(gs.Player.Skills[formula.SkillTrader], gs.Player.CrewMercs(), formula.SkillTrader, 0)
}

func Buy(gs *game.GameState, goodIdx int, qty int) TransactionResult {
	if goodIdx < 0 || goodIdx >= game.NumGoods {
		return TransactionResult{Message: "Invalid good."}
	}
	if qty <= 0 {
		return TransactionResult{Message: "Invalid quantity."}
	}

	sysState := &gs.Systems[gs.CurrentSystemID]
	basePrice := sysState.Prices[goodIdx]
	if basePrice < 0 {
		return TransactionResult{Message: "Good not available in this market."}
	}

	traderSkill := effectiveTraderSkill(gs)
	discount := traderSkill
	if discount > 10 {
		discount = 10
	}
	price := basePrice * (100 - discount) / 100
	if price < 1 {
		price = 1
	}

	totalCost := price * qty
	if gs.Player.Credits < totalCost {
		return TransactionResult{Message: "Not enough credits."}
	}

	dp := &game.GameDataProvider{Data: gs.Data}
	if gs.Player.FreeCargo(dp) < qty {
		return TransactionResult{Message: "Not enough cargo space."}
	}

	gs.Player.Credits -= totalCost
	gs.Player.CargoCost[goodIdx] += totalCost
	gs.Player.Cargo[goodIdx] += qty

	goodName := gs.Data.Goods[goodIdx].Name
	msg := fmt.Sprintf("Bought %d %s for %d cr", qty, goodName, totalCost)
	if discount > 0 {
		msg += fmt.Sprintf(" (%d%% trader discount)", discount)
	}
	return TransactionResult{
		Success:    true,
		Message:    msg,
		TotalPrice: totalCost,
	}
}

func Sell(gs *game.GameState, goodIdx int, qty int) TransactionResult {
	if goodIdx < 0 || goodIdx >= game.NumGoods {
		return TransactionResult{Message: "Invalid good."}
	}
	if qty <= 0 {
		return TransactionResult{Message: "Invalid quantity."}
	}

	if gs.Player.Cargo[goodIdx] < qty {
		return TransactionResult{Message: "Not enough goods to sell."}
	}

	sysState := &gs.Systems[gs.CurrentSystemID]
	basePrice := sysState.Prices[goodIdx]
	if basePrice < 0 {
		return TransactionResult{Message: "Market does not buy this good here."}
	}

	traderSkill := effectiveTraderSkill(gs)
	bonus := traderSkill
	if bonus > 10 {
		bonus = 10
	}
	price := basePrice * (100 + bonus) / 100

	totalPrice := price * qty

	costBasis := 0
	if gs.Player.Cargo[goodIdx] > 0 {
		costBasis = gs.Player.CargoCost[goodIdx] * qty / gs.Player.Cargo[goodIdx]
	}

	gs.Player.Credits += totalPrice
	gs.Player.CargoCost[goodIdx] -= costBasis
	gs.Player.Cargo[goodIdx] -= qty
	if gs.Player.Cargo[goodIdx] == 0 {
		gs.Player.CargoCost[goodIdx] = 0
	}

	profit := totalPrice - costBasis
	goodName := gs.Data.Goods[goodIdx].Name
	msg := fmt.Sprintf("Sold %d %s for %d cr", qty, goodName, totalPrice)
	if costBasis > 0 {
		if profit >= 0 {
			msg += fmt.Sprintf(" (profit: +%d)", profit)
		} else {
			msg += fmt.Sprintf(" (loss: %d)", profit)
		}
	}
	if bonus > 0 {
		msg += fmt.Sprintf(" (%d%% trader bonus)", bonus)
	}
	return TransactionResult{
		Success:    true,
		Message:    msg,
		TotalPrice: totalPrice,
	}
}
