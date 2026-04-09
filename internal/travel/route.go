package travel

import (
	"math"

	"github.com/the4ofus/spacetrader-tui/internal/formula"
	"github.com/the4ofus/spacetrader-tui/internal/game"
)

type RouteHop struct {
	SystemIdx   int
	FuelCost    int
	RefuelCost  int
	IsWormhole  bool
	WormholeFee int
	Distance    float64
}

type Route struct {
	Hops          []RouteHop
	TotalFuel     int
	TotalRefuel   int
	TotalWormhole int
	Reachable     bool
}

func FindRoute(gs *game.GameState, destIdx int) Route {
	return FindRouteFrom(gs, gs.CurrentSystemID, destIdx)
}

func FindRouteFrom(gs *game.GameState, originIdx, destIdx int) Route {
	systems := gs.Data.Systems
	n := len(systems)
	shipRange := gs.EffectiveRange()
	shipDef := gs.Data.Ships[gs.Player.Ship.TypeID]
	origin := originIdx

	if origin == destIdx {
		return Route{
			Hops:      []RouteHop{{SystemIdx: origin}},
			Reachable: true,
		}
	}

	dist := make([]int, n)
	prev := make([]int, n)
	viaWH := make([]bool, n)
	visited := make([]bool, n)

	for i := range dist {
		dist[i] = math.MaxInt32
		prev[i] = -1
	}
	dist[origin] = 0

	for {
		u := -1
		for i := 0; i < n; i++ {
			if !visited[i] && (u == -1 || dist[i] < dist[u]) {
				u = i
			}
		}
		if u == -1 || dist[u] == math.MaxInt32 {
			break
		}
		if u == destIdx {
			break
		}
		visited[u] = true

		for v := 0; v < n; v++ {
			if v == u || visited[v] {
				continue
			}
			d := formula.Distance(systems[u].X, systems[u].Y, systems[v].X, systems[v].Y)
			fuelCost := int(math.Ceil(d))
			if fuelCost <= shipRange {
				newDist := dist[u] + fuelCost
				if newDist < dist[v] {
					dist[v] = newDist
					prev[v] = u
					viaWH[v] = false
				}
			}
		}

		if whDest, ok := game.WormholeDestination(gs, u); ok {
			if !visited[whDest] && dist[u] < dist[whDest] {
				dist[whDest] = dist[u]
				prev[whDest] = u
				viaWH[whDest] = true
			}
		}
	}

	if dist[destIdx] == math.MaxInt32 {
		return Route{Reachable: false}
	}

	var path []int
	var whFlags []bool
	for at := destIdx; at != -1; at = prev[at] {
		path = append(path, at)
		whFlags = append(whFlags, viaWH[at])
	}
	for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
		path[i], path[j] = path[j], path[i]
		whFlags[i], whFlags[j] = whFlags[j], whFlags[i]
	}

	wormholeFee := game.WormholeTax(gs)

	route := Route{Reachable: true}
	for i, sysIdx := range path {
		hop := RouteHop{SystemIdx: sysIdx}
		if i > 0 {
			hop.IsWormhole = whFlags[i]
			if hop.IsWormhole {
				hop.FuelCost = 0
				hop.Distance = 0
				hop.WormholeFee = wormholeFee
				hop.RefuelCost = 0
				route.TotalWormhole += wormholeFee
			} else {
				prevSys := systems[path[i-1]]
				curSys := systems[sysIdx]
				d := formula.Distance(prevSys.X, prevSys.Y, curSys.X, curSys.Y)
				hop.Distance = d
				hop.FuelCost = int(math.Ceil(d))
				hop.RefuelCost = hop.FuelCost * shipDef.FuelCost
				route.TotalFuel += hop.FuelCost
				route.TotalRefuel += hop.RefuelCost
			}
		}
		route.Hops = append(route.Hops, hop)
	}

	return route
}
