package travel

import (
	"sort"

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
	NoFromInfo bool
	NoToInfo   bool
	FromStale  bool
	ToStale    bool
}

const maxTradesPerHop = 3

func AnalyzeRouteTrades(gs *game.GameState, route Route) []HopTradeInfo {
	if len(route.Hops) < 2 {
		return nil
	}

	result := make([]HopTradeInfo, len(route.Hops)-1)
	for i := 0; i < len(route.Hops)-1; i++ {
		fromIdx := route.Hops[i].SystemIdx
		toIdx := route.Hops[i+1].SystemIdx
		info := HopTradeInfo{FromIdx: fromIdx, ToIdx: toIdx}

		_, hasFrom := gs.GetTradeInfo(fromIdx)
		_, hasTo := gs.GetTradeInfo(toIdx)
		fromIsCurrent := fromIdx == gs.CurrentSystemID

		if !hasFrom && !fromIsCurrent {
			info.NoFromInfo = true
		}
		if !hasTo {
			info.NoToInfo = true
		}

		if info.NoFromInfo || info.NoToInfo {
			result[i] = info
			continue
		}

		if !fromIsCurrent {
			stale, _ := gs.IsTradeInfoStale(fromIdx)
			info.FromStale = stale
		}
		if hasTo {
			stale, _ := gs.IsTradeInfoStale(toIdx)
			info.ToStale = stale
		}

		for g, good := range gs.Data.Goods {
			buyPrice := gs.Systems[fromIdx].Prices[g]
			sellPrice := gs.Systems[toIdx].Prices[g]
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
