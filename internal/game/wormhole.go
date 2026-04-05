package game

type Wormhole struct {
	SystemA int `json:"system_a"`
	SystemB int `json:"system_b"`
}

func GenerateWormholes(gs *GameState) {
	numWormholes := 3 + gs.Rand.Intn(3)
	used := map[int]bool{}

	for len(gs.Wormholes) < numWormholes {
		a := gs.Rand.Intn(len(gs.Data.Systems))
		b := gs.Rand.Intn(len(gs.Data.Systems))
		if a == b || used[a] || used[b] {
			continue
		}
		used[a] = true
		used[b] = true
		gs.Wormholes = append(gs.Wormholes, Wormhole{SystemA: a, SystemB: b})
	}
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

func TravelWormhole(gs *GameState) (bool, string) {
	dest, ok := WormholeDestination(gs, gs.CurrentSystemID)
	if !ok {
		return false, "No wormhole at this system."
	}

	gs.CurrentSystemID = dest
	gs.Systems[dest].Visited = true
	gs.Day++

	destName := gs.Data.Systems[dest].Name
	return true, "Traveled through wormhole to " + destName + "!"
}
