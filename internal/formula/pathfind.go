package formula

import "math"

type WormholePair struct {
	A, B int
}

func ShortestPathHops(systems [][2]int, shipRange int, wormholes []WormholePair, from, to int) int {
	n := len(systems)
	if from == to {
		return 0
	}

	dist := make([]int, n)
	for i := range dist {
		dist[i] = math.MaxInt32
	}
	dist[from] = 0
	visited := make([]bool, n)
	rangeF := float64(shipRange)

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
		if u == to {
			return dist[to]
		}
		visited[u] = true

		for v := 0; v < n; v++ {
			if v == u || visited[v] {
				continue
			}
			d := Distance(systems[u][0], systems[u][1], systems[v][0], systems[v][1])
			if math.Ceil(d) <= rangeF {
				newDist := dist[u] + 1
				if newDist < dist[v] {
					dist[v] = newDist
				}
			}
		}

		for _, wh := range wormholes {
			dest := -1
			if wh.A == u {
				dest = wh.B
			} else if wh.B == u {
				dest = wh.A
			}
			if dest >= 0 && !visited[dest] {
				newDist := dist[u] + 1
				if newDist < dist[dest] {
					dist[dest] = newDist
				}
			}
		}
	}

	return -1
}
