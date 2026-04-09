package travel_test

import (
	"testing"

	"github.com/the4ofus/spacetrader-tui/internal/formula"
	"github.com/the4ofus/spacetrader-tui/internal/travel"
)

func TestRouteDirectHop(t *testing.T) {
	gs := newTestGame(t)
	reachable := travel.ReachableSystems(gs)
	if len(reachable) == 0 {
		t.Skip("no reachable systems")
	}

	destIdx := reachable[0].Index
	route := travel.FindRoute(gs, destIdx)
	if !route.Reachable {
		t.Fatal("direct hop should be reachable")
	}
	if len(route.Hops) != 2 {
		t.Errorf("direct hop should have 2 hops (origin + dest), got %d", len(route.Hops))
	}
	if route.Hops[0].SystemIdx != gs.CurrentSystemID {
		t.Error("first hop should be current system")
	}
	if route.Hops[1].SystemIdx != destIdx {
		t.Error("second hop should be destination")
	}
}

func TestRouteSameSystem(t *testing.T) {
	gs := newTestGame(t)
	route := travel.FindRoute(gs, gs.CurrentSystemID)
	if !route.Reachable {
		t.Fatal("route to self should be reachable")
	}
	if len(route.Hops) != 1 {
		t.Errorf("route to self should have 1 hop, got %d", len(route.Hops))
	}
}

func TestRouteMultiHop(t *testing.T) {
	gs := newTestGame(t)
	shipRange := float64(gs.EffectiveRange())
	cur := gs.Data.Systems[gs.CurrentSystemID]

	farIdx := -1
	for i, sys := range gs.Data.Systems {
		if i == gs.CurrentSystemID {
			continue
		}
		dist := formula.Distance(cur.X, cur.Y, sys.X, sys.Y)
		if dist > shipRange && dist < shipRange*4 {
			farIdx = i
			break
		}
	}
	if farIdx == -1 {
		t.Skip("no system found requiring multi-hop")
	}

	route := travel.FindRoute(gs, farIdx)
	if !route.Reachable {
		t.Fatalf("system %d should be reachable via multi-hop", farIdx)
	}
	if len(route.Hops) < 3 {
		t.Errorf("expected at least 3 hops for a far system, got %d", len(route.Hops))
	}

	if route.TotalFuel <= 0 {
		t.Error("total fuel should be positive for multi-hop route")
	}
	if route.TotalRefuel <= 0 {
		t.Error("total refuel cost should be positive")
	}

	for i := 1; i < len(route.Hops); i++ {
		hop := route.Hops[i]
		if !hop.IsWormhole && hop.FuelCost <= 0 {
			t.Errorf("hop %d: non-wormhole should have positive fuel cost", i)
		}
	}
}

func TestRouteTradeAnalysis(t *testing.T) {
	gs := newTestGame(t)

	farIdx := -1
	cur := gs.Data.Systems[gs.CurrentSystemID]
	shipRange := float64(gs.EffectiveRange())
	for i, sys := range gs.Data.Systems {
		if i == gs.CurrentSystemID {
			continue
		}
		dist := formula.Distance(cur.X, cur.Y, sys.X, sys.Y)
		if dist > shipRange {
			farIdx = i
			break
		}
	}
	if farIdx == -1 {
		t.Skip("no multi-hop system found")
	}

	route := travel.FindRoute(gs, farIdx)
	if !route.Reachable {
		t.Skip("route not reachable")
	}

	trades := travel.AnalyzeRouteTrades(gs, route)
	if len(trades) != len(route.Hops)-1 {
		t.Errorf("expected %d trade infos, got %d", len(route.Hops)-1, len(trades))
	}

	for i, info := range trades {
		for _, tr := range info.Trades {
			if tr.Profit <= 0 {
				t.Errorf("hop %d: trade %s has non-positive profit %d", i, tr.GoodName, tr.Profit)
			}
			if tr.BuyPrice <= 0 || tr.SellPrice <= 0 {
				t.Errorf("hop %d: trade %s has invalid prices buy=%d sell=%d", i, tr.GoodName, tr.BuyPrice, tr.SellPrice)
			}
		}
		if len(info.Trades) > 3 {
			t.Errorf("hop %d: should have at most 3 trades, got %d", i, len(info.Trades))
		}
	}
}
