package travel

import (
	"sort"

	"github.com/the4ofus/spacetrader-tui/internal/formula"
	"github.com/the4ofus/spacetrader-tui/internal/game"
)

type HopTrade struct {
	GoodIdx   int
	GoodName  string
	BuyPrice  int
	SellPrice int
	Profit    int
}

type HopTradeInfo struct {
	FromIdx    int
	ToIdx      int
	Trades     []HopTrade
	OutOfRange bool
}

const maxTradesPerHop = 3
const maxPriceDataRange = 50.0

func AnalyzeRouteTrades(gs *game.GameState, route Route) []HopTradeInfo {
	if len(route.Hops) < 2 {
		return nil
	}

	cur := gs.Data.Systems[gs.CurrentSystemID]

	result := make([]HopTradeInfo, len(route.Hops)-1)
	for i := 0; i < len(route.Hops)-1; i++ {
		fromIdx := route.Hops[i].SystemIdx
		toIdx := route.Hops[i+1].SystemIdx
		info := HopTradeInfo{FromIdx: fromIdx, ToIdx: toIdx}

		fromSys := gs.Data.Systems[fromIdx]
		toSys := gs.Data.Systems[toIdx]
		fromDist := formula.Distance(cur.X, cur.Y, fromSys.X, fromSys.Y)
		toDist := formula.Distance(cur.X, cur.Y, toSys.X, toSys.Y)

		if fromDist > maxPriceDataRange || toDist > maxPriceDataRange {
			info.OutOfRange = true
			result[i] = info
			continue
		}

		for g, good := range gs.Data.Goods {
			buyPrice := game.BuyPriceAt(gs, fromIdx, g)
			sellPrice := game.SellPriceAt(gs, toIdx, g)
			if buyPrice <= 0 || sellPrice <= 0 {
				continue
			}
			profit := sellPrice - buyPrice
			if profit <= 0 {
				continue
			}
			info.Trades = append(info.Trades, HopTrade{
				GoodIdx:   g,
				GoodName:  good.Name,
				BuyPrice:  buyPrice,
				SellPrice: sellPrice,
				Profit:    profit,
			})
		}

		sort.Slice(info.Trades, func(a, b int) bool {
			return info.Trades[a].Profit > info.Trades[b].Profit
		})
		if len(info.Trades) > maxTradesPerHop {
			info.Trades = info.Trades[:maxTradesPerHop]
		}

		result[i] = info
	}
	return result
}
