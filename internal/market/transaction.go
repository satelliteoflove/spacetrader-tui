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
	return game.EffectivePlayerSkill(gs, formula.SkillTrader)
}

func SellPrice(gs *game.GameState, goodIdx int) int {
	basePrice := gs.Systems[gs.CurrentSystemID].Prices[goodIdx]
	if basePrice < 0 {
		return -1
	}
	if gs.Player.PoliceRecord < -5 {
		basePrice = basePrice * 90 / 100
	}
	if basePrice < 1 {
		basePrice = 1
	}
	return basePrice
}

func BuyPrice(gs *game.GameState, goodIdx int) int {
	sellPrice := SellPrice(gs, goodIdx)
	if sellPrice < 0 {
		return -1
	}

	base := sellPrice
	if gs.Player.PoliceRecord < -5 {
		base = base * 100 / 90
	}

	traderSkill := effectiveTraderSkill(gs)
	if traderSkill > 10 {
		traderSkill = 10
	}
	buyPrice := base * (103 + (10 - traderSkill)) / 100

	if buyPrice <= sellPrice {
		buyPrice = sellPrice + 1
	}
	return buyPrice
}

func Buy(gs *game.GameState, goodIdx int, qty int) TransactionResult {
	if goodIdx < 0 || goodIdx >= game.NumGoods {
		return TransactionResult{Message: "Invalid good."}
	}
	if qty <= 0 {
		return TransactionResult{Message: "Invalid quantity."}
	}

	price := BuyPrice(gs, goodIdx)
	if price < 0 {
		return TransactionResult{Message: "Good not available in this market."}
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
	return TransactionResult{
		Success:    true,
		Message:    fmt.Sprintf("Bought %d %s for %d cr (%d cr/unit)", qty, goodName, totalCost, price),
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

	price := SellPrice(gs, goodIdx)
	if price < 0 {
		return TransactionResult{Message: "Market does not buy this good here."}
	}

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
	return TransactionResult{
		Success:    true,
		Message:    msg,
		TotalPrice: totalPrice,
	}
}
