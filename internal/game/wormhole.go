package game

import (
	"math"

	"github.com/the4ofus/spacetrader-tui/internal/gamedata"
)

const NumWormholes = 6

type Wormhole struct {
	SystemA int `json:"system_a"`
	SystemB int `json:"system_b"`
}

func GenerateWormholes(gs *GameState) {
	systems := gs.Data.Systems
	nodes := pickWormholeNodes(gs, systems)

	gs.Wormholes = make([]Wormhole, NumWormholes)
	for i := 0; i < NumWormholes; i++ {
		next := (i + 1) % NumWormholes
		gs.Wormholes[i] = Wormhole{SystemA: nodes[i], SystemB: nodes[next]}
	}
}

func pickWormholeNodes(gs *GameState, systems []gamedata.SystemDef) []int {
	first := gs.Rand.Intn(len(systems))
	nodes := []int{first}

	used := map[int]bool{first: true}

	for len(nodes) < NumWormholes {
		bestIdx := -1
		bestMinDist := -1.0

		for i := range systems {
			if used[i] {
				continue
			}
			minD := math.MaxFloat64
			for _, n := range nodes {
				d := systemDist(systems[i], systems[n])
				if d < minD {
					minD = d
				}
			}
			if minD > bestMinDist {
				bestMinDist = minD
				bestIdx = i
			}
		}

		if bestIdx >= 0 {
			nodes = append(nodes, bestIdx)
			used[bestIdx] = true
		}
	}

	return nodes
}

func systemDist(a, b gamedata.SystemDef) float64 {
	dx := float64(a.X - b.X)
	dy := float64(a.Y - b.Y)
	return math.Sqrt(dx*dx + dy*dy)
}

func WormholeDestination(gs *GameState, fromSystem int) (int, bool) {
	for _, wh := range gs.Wormholes {
		if wh.SystemA == fromSystem {
			return wh.SystemB, true
		}
		if wh.SystemB == fromSystem {
			return wh.SystemA, true
		}
	}
	return 0, false
}

func WormholeTax(gs *GameState) int {
	shipDef := gs.Data.Ships[gs.Player.Ship.TypeID]
	return shipDef.FuelCost * 25
}

func TravelWormhole(gs *GameState) (bool, string) {
	dest, ok := WormholeDestination(gs, gs.CurrentSystemID)
	if !ok {
		return false, "No wormhole at this system."
	}

	tax := WormholeTax(gs)
	if gs.Player.Credits < tax {
		return false, "Insufficient credits for wormhole transit fee."
	}

	gs.Player.Credits -= tax
	gs.CurrentSystemID = dest
	gs.Systems[dest].Visited = true
	gs.Day++
	gs.CaptureTradeInfo(dest)
	gs.RecordSnapshot()

	destName := gs.Data.Systems[dest].Name
	return true, "Traveled through wormhole to " + destName + "!"
}

func IsWormholeSystem(gs *GameState, sysIdx int) bool {
	for _, wh := range gs.Wormholes {
		if wh.SystemA == sysIdx || wh.SystemB == sysIdx {
			return true
		}
	}
	return false
}
