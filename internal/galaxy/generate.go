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
	MinDistance      = 5
	CloseDistance    = 13
	MaxShipRange    = 17
	MaxPlaceRetries = 1000
)

func Generate(seed int64) []gamedata.SystemDef {
	rng := rand.New(rand.NewSource(seed))
	names := shuffledNames(rng)
	systems := make([]gamedata.SystemDef, NumSystems)

	coords := placeCoordinates(rng)

	for i := 0; i < NumSystems; i++ {
		tech := gamedata.TechLevel(rng.Intn(int(gamedata.NumTechLevels)))
		systems[i] = gamedata.SystemDef{
			ID:              i,
			Name:            names[i],
			X:               coords[i][0],
			Y:               coords[i][1],
			TechLevel:       tech,
			PoliticalSystem: gamedata.PoliticalSystem(rng.Intn(int(gamedata.NumPoliticalSystems))),
			Resource:        randomResource(rng, tech),
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
	for _, n := range allSystemNames() {
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

const (
	NumChains    = 4
	ChainStepMin = 7
	ChainStepMax = 9
	ForkChance   = 15
	DriftMax     = 0.6
	EdgeMargin   = 5
)

type chainHead struct {
	x, y  float64
	angle float64
}

func placeCoordinates(rng *rand.Rand) [][2]int {
	coords := make([][2]int, 0, NumSystems)

	seeds := spreadSeeds(rng, NumChains)
	var heads []chainHead
	for _, s := range seeds {
		coords = append(coords, s)
		heads = append(heads, chainHead{
			x:     float64(s[0]),
			y:     float64(s[1]),
			angle: rng.Float64() * 2 * math.Pi,
		})
	}

	for len(coords) < NumSystems && len(heads) > 0 {
		idx := rng.Intn(len(heads))
		h := &heads[idx]

		stepDist := float64(ChainStepMin) + rng.Float64()*float64(ChainStepMax-ChainStepMin)
		h.angle += (rng.Float64()*2 - 1) * DriftMax

		nx := h.x + math.Cos(h.angle)*stepDist
		ny := h.y + math.Sin(h.angle)*stepDist

		if nx < EdgeMargin || nx >= float64(GalaxyWidth-EdgeMargin) {
			h.angle = math.Pi - h.angle
			nx = h.x + math.Cos(h.angle)*stepDist
		}
		if ny < EdgeMargin || ny >= float64(GalaxyHeight-EdgeMargin) {
			h.angle = -h.angle
			ny = h.y + math.Sin(h.angle)*stepDist
		}

		nx = math.Max(1, math.Min(float64(GalaxyWidth-2), nx))
		ny = math.Max(1, math.Min(float64(GalaxyHeight-2), ny))

		ix, iy := int(nx), int(ny)
		if tooClose(coords, ix, iy) {
			heads = append(heads[:idx], heads[idx+1:]...)
			continue
		}

		coords = append(coords, [2]int{ix, iy})
		h.x, h.y = nx, ny

		if rng.Intn(100) < ForkChance && len(heads) < 10 {
			forkAngle := h.angle + (math.Pi/3 + rng.Float64()*math.Pi/3)
			if rng.Intn(2) == 0 {
				forkAngle = h.angle - (math.Pi/3 + rng.Float64()*math.Pi/3)
			}
			heads = append(heads, chainHead{x: nx, y: ny, angle: forkAngle})
		}
	}

	for len(coords) < NumSystems {
		placed := false
		for attempt := 0; attempt < MaxPlaceRetries; attempt++ {
			x := rng.Intn(GalaxyWidth)
			y := rng.Intn(GalaxyHeight)
			if !tooClose(coords, x, y) {
				coords = append(coords, [2]int{x, y})
				placed = true
				break
			}
		}
		if !placed {
			coords = append(coords, forcePlace(rng, coords))
		}
	}

	ensureConnectivity(rng, coords)
	return coords
}

func spreadSeeds(rng *rand.Rand, n int) [][2]int {
	seeds := make([][2]int, 0, n)
	for i := 0; i < n; i++ {
		for attempt := 0; attempt < 200; attempt++ {
			x := EdgeMargin + rng.Intn(GalaxyWidth-2*EdgeMargin)
			y := EdgeMargin + rng.Intn(GalaxyHeight-2*EdgeMargin)
			tooNear := false
			for _, s := range seeds {
				if dist(s[0], s[1], x, y) < float64(GalaxyWidth)/float64(n) {
					tooNear = true
					break
				}
			}
			if !tooNear {
				seeds = append(seeds, [2]int{x, y})
				break
			}
		}
		if len(seeds) <= i {
			seeds = append(seeds, [2]int{
				EdgeMargin + rng.Intn(GalaxyWidth-2*EdgeMargin),
				EdgeMargin + rng.Intn(GalaxyHeight-2*EdgeMargin),
			})
		}
	}
	return seeds
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
	for attempt := 0; attempt < 200; attempt++ {
		components := connectedComponents(coords, MaxShipRange)
		if len(components) <= 1 {
			return
		}
		main := components[0]
		for _, island := range components[1:] {
			bridgeNode := island[0]
			bestTarget := -1
			bestDist := math.MaxFloat64
			for _, mi := range main {
				d := dist(coords[bridgeNode][0], coords[bridgeNode][1], coords[mi][0], coords[mi][1])
				if d < bestDist {
					bestDist = d
					bestTarget = mi
				}
			}
			if bestTarget >= 0 {
				nudgeCloser(rng, coords, bridgeNode, bestTarget)
			}
			main = append(main, island...)
		}
	}
}

func connectedComponents(coords [][2]int, maxRange int) [][]int {
	n := len(coords)
	visited := make([]bool, n)
	rangeF := float64(maxRange)
	var components [][]int

	for start := 0; start < n; start++ {
		if visited[start] {
			continue
		}
		var component []int
		queue := []int{start}
		visited[start] = true
		for len(queue) > 0 {
			cur := queue[0]
			queue = queue[1:]
			component = append(component, cur)
			for j := 0; j < n; j++ {
				if visited[j] {
					continue
				}
				if math.Ceil(dist(coords[cur][0], coords[cur][1], coords[j][0], coords[j][1])) <= rangeF {
					visited[j] = true
					queue = append(queue, j)
				}
			}
		}
		components = append(components, component)
	}
	return components
}

func nudgeCloser(rng *rand.Rand, coords [][2]int, isolated, target int) {
	tx, ty := coords[target][0], coords[target][1]
	ix, iy := coords[isolated][0], coords[isolated][1]

	d := dist(ix, iy, tx, ty)
	if math.Ceil(d) <= float64(MaxShipRange) {
		return
	}

	dx := float64(tx - ix)
	dy := float64(ty - iy)
	moveRatio := (d - float64(MaxShipRange) + 2) / d
	newX := ix + int(math.Round(dx*moveRatio))
	newY := iy + int(math.Round(dy*moveRatio))

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

var validResources = []struct {
	Resource gamedata.Resource
	MinTech  gamedata.TechLevel
}{
	{gamedata.ResourceMineralRich, gamedata.TechMedieval},
	{gamedata.ResourceMineralPoor, gamedata.TechMedieval},
	{gamedata.ResourceDesert, gamedata.TechPreAgricultural},
	{gamedata.ResourceSweetOceans, gamedata.TechPreAgricultural},
	{gamedata.ResourceRichSoil, gamedata.TechAgricultural},
	{gamedata.ResourcePoorSoil, gamedata.TechAgricultural},
	{gamedata.ResourceRichFauna, gamedata.TechPreAgricultural},
	{gamedata.ResourceLifeless, gamedata.TechPreAgricultural},
	{gamedata.ResourceWeirdMushrooms, gamedata.TechIndustrial},
	{gamedata.ResourceSpecialHerbs, gamedata.TechEarlyIndustrial},
	{gamedata.ResourceArtistic, gamedata.TechRenaissance},
	{gamedata.ResourceWarlike, gamedata.TechRenaissance},
}

func randomResource(rng *rand.Rand, tech gamedata.TechLevel) gamedata.Resource {
	if rng.Intn(100) < 60 {
		return gamedata.ResourceNone
	}
	var candidates []gamedata.Resource
	for _, vr := range validResources {
		if tech >= vr.MinTech {
			candidates = append(candidates, vr.Resource)
		}
	}
	if len(candidates) == 0 {
		return gamedata.ResourceNone
	}
	return candidates[rng.Intn(len(candidates))]
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
