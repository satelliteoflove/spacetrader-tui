package galaxy

import (
	"math"
	"math/rand"

	"github.com/the4ofus/spacetrader-tui/internal/gamedata"
)

const (
	GalaxyWidth    = 150
	GalaxyHeight   = 110
	NumSystems     = 120
	MinDistance     = 6
	CloseDistance   = 13
	MaxPlaceRetries = 1000
)

func Generate(seed int64) []gamedata.SystemDef {
	rng := rand.New(rand.NewSource(seed))
	names := shuffledNames(rng)
	systems := make([]gamedata.SystemDef, NumSystems)

	coords := placeCoordinates(rng)

	for i := 0; i < NumSystems; i++ {
		systems[i] = gamedata.SystemDef{
			ID:              i,
			Name:            names[i],
			X:               coords[i][0],
			Y:               coords[i][1],
			TechLevel:       gamedata.TechLevel(rng.Intn(int(gamedata.NumTechLevels))),
			PoliticalSystem: gamedata.PoliticalSystem(rng.Intn(int(gamedata.NumPoliticalSystems))),
			Resource:        randomResource(rng),
			Size:            gamedata.SystemSize(rng.Intn(int(gamedata.NumSystemSizes))),
		}
	}

	return systems
}

func shuffledNames(rng *rand.Rand) []string {
	critical := make(map[string]bool, len(questCriticalNames))
	for _, n := range questCriticalNames {
		critical[n] = true
	}

	result := make([]string, 0, NumSystems)
	result = append(result, questCriticalNames...)

	var others []string
	for _, n := range systemNamePool {
		if !critical[n] {
			others = append(others, n)
		}
	}
	rng.Shuffle(len(others), func(i, j int) {
		others[i], others[j] = others[j], others[i]
	})

	result = append(result, others...)
	if len(result) > NumSystems {
		result = result[:NumSystems]
	}

	rng.Shuffle(len(result), func(i, j int) {
		result[i], result[j] = result[j], result[i]
	})

	return result
}

func placeCoordinates(rng *rand.Rand) [][2]int {
	coords := make([][2]int, 0, NumSystems)

	for len(coords) < NumSystems {
		placed := false
		for attempt := 0; attempt < MaxPlaceRetries; attempt++ {
			x := rng.Intn(GalaxyWidth)
			y := rng.Intn(GalaxyHeight)

			if tooClose(coords, x, y) {
				continue
			}

			coords = append(coords, [2]int{x, y})
			placed = true
			break
		}
		if !placed {
			coords = append(coords, forcePlace(rng, coords))
		}
	}

	ensureConnectivity(rng, coords)
	return coords
}

func tooClose(coords [][2]int, x, y int) bool {
	for _, c := range coords {
		if dist(c[0], c[1], x, y) < MinDistance {
			return true
		}
	}
	return false
}

func forcePlace(rng *rand.Rand, existing [][2]int) [2]int {
	bestX, bestY := 0, 0
	bestMinDist := 0.0

	for attempt := 0; attempt < 500; attempt++ {
		x := rng.Intn(GalaxyWidth)
		y := rng.Intn(GalaxyHeight)
		minD := math.MaxFloat64
		for _, c := range existing {
			d := dist(c[0], c[1], x, y)
			if d < minD {
				minD = d
			}
		}
		if minD > bestMinDist {
			bestMinDist = minD
			bestX, bestY = x, y
		}
	}
	return [2]int{bestX, bestY}
}

func ensureConnectivity(rng *rand.Rand, coords [][2]int) {
	for i := range coords {
		hasNeighbor := false
		for j := range coords {
			if i != j && dist(coords[i][0], coords[i][1], coords[j][0], coords[j][1]) <= CloseDistance {
				hasNeighbor = true
				break
			}
		}
		if !hasNeighbor {
			nearest := -1
			nearestDist := math.MaxFloat64
			for j := range coords {
				if i == j {
					continue
				}
				d := dist(coords[i][0], coords[i][1], coords[j][0], coords[j][1])
				if d < nearestDist {
					nearestDist = d
					nearest = j
				}
			}
			if nearest >= 0 {
				nudgeCloser(rng, coords, i, nearest)
			}
		}
	}
}

func nudgeCloser(rng *rand.Rand, coords [][2]int, isolated, target int) {
	tx, ty := coords[target][0], coords[target][1]
	ix, iy := coords[isolated][0], coords[isolated][1]

	dx := tx - ix
	dy := ty - iy
	d := dist(ix, iy, tx, ty)
	if d <= CloseDistance {
		return
	}

	moveRatio := (d - CloseDistance + 1) / d
	newX := ix + int(float64(dx)*moveRatio)
	newY := iy + int(float64(dy)*moveRatio)

	if newX < 0 {
		newX = 0
	}
	if newX >= GalaxyWidth {
		newX = GalaxyWidth - 1
	}
	if newY < 0 {
		newY = 0
	}
	if newY >= GalaxyHeight {
		newY = GalaxyHeight - 1
	}

	coords[isolated] = [2]int{newX, newY}
}

func randomResource(rng *rand.Rand) gamedata.Resource {
	if rng.Intn(100) < 60 {
		return gamedata.ResourceNone
	}
	return gamedata.Resource(1 + rng.Intn(int(gamedata.NumResources)-1))
}

func dist(x1, y1, x2, y2 int) float64 {
	dx := float64(x2 - x1)
	dy := float64(y2 - y1)
	return math.Sqrt(dx*dx + dy*dy)
}

func FindSystemByName(systems []gamedata.SystemDef, name string) int {
	for i, s := range systems {
		if s.Name == name {
			return i
		}
	}
	return -1
}
