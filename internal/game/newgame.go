package game

import (
	"math/rand"
	"time"

	"github.com/the4ofus/spacetrader-tui/internal/formula"
	"github.com/the4ofus/spacetrader-tui/internal/galaxy"
	"github.com/the4ofus/spacetrader-tui/internal/gamedata"
)

func NewGame(data *gamedata.GameData, name string, skills [formula.NumSkills]int, diff gamedata.Difficulty) *GameState {
	seed := time.Now().UnixNano()
	return NewGameWithSeed(data, name, skills, diff, seed)
}

func NewGameWithSeed(data *gamedata.GameData, name string, skills [formula.NumSkills]int, diff gamedata.Difficulty, seed int64) *GameState {
	rng := rand.New(rand.NewSource(seed))

	data.Systems = galaxy.Generate(seed)

	gnatDef := data.Ships[int(formula.StartingShip)]

	gs := &GameState{
		Player: Player{
			Name:    name,
			Credits: formula.StartingCredits,
			Skills:  skills,
			Ship: Ship{
				TypeID:  int(formula.StartingShip),
				Hull:    gnatDef.Hull,
				Fuel:    gnatDef.Range,
				Weapons: []int{0},
				Shields: []int{},
				Gadgets: []int{},
			},
		},
		Systems:     make([]SystemState, len(data.Systems)),
		Day:         1,
		Difficulty:  diff,
		EndStatus:   StatusPlaying,
		SaveVersion: CurrentSaveVersion,
		Seed:        seed,
		Rand:        rng,
		Data:        data,
	}

	startIdx := pickStartingSystem(gs)
	gs.CurrentSystemID = startIdx
	gs.Systems[startIdx].Visited = true

	sysCoords := make([][2]int, len(data.Systems))
	for i, s := range data.Systems {
		sysCoords[i] = [2]int{s.X, s.Y}
	}
	gs.Mercenaries = GenerateMercenaries(rng, len(data.Systems), startIdx, sysCoords, gs.EffectiveRange())

	initializeMarkets(gs)
	GenerateWormholes(gs)
	GenerateEvents(gs)
	ensureNearbyEvent(gs)

	if diff == gamedata.DiffBeginner {
		gs.Player.Credits += 1000
	}

	return gs
}

func pickStartingSystem(gs *GameState) int {
	shipRange := float64(gs.Data.Ships[gs.Player.Ship.TypeID].Range)
	minNeighbors := 3

	candidates := []int{}
	for i, sys := range gs.Data.Systems {
		if sys.TechLevel >= gamedata.TechAgricultural && sys.PoliticalSystem != gamedata.PolAnarchy {
			neighbors := 0
			for j, other := range gs.Data.Systems {
				if i != j && formula.Distance(sys.X, sys.Y, other.X, other.Y) <= shipRange {
					neighbors++
				}
			}
			if neighbors >= minNeighbors {
				candidates = append(candidates, i)
			}
		}
	}
	if len(candidates) == 0 {
		for i, sys := range gs.Data.Systems {
			if sys.TechLevel >= gamedata.TechAgricultural {
				candidates = append(candidates, i)
			}
		}
	}
	if len(candidates) == 0 {
		return 0
	}
	return candidates[gs.Rand.Intn(len(candidates))]
}

func NewStartingShip(data *gamedata.GameData) Ship {
	fleaDef := data.Ships[ShipFlea]
	return Ship{
		TypeID:  ShipFlea,
		Hull:    fleaDef.Hull,
		Fuel:    fleaDef.Range,
		Weapons: []int{},
		Shields: []int{},
		Gadgets: []int{},
	}
}

func ensureNearbyEvent(gs *GameState) {
	cur := gs.Data.Systems[gs.CurrentSystemID]
	for i, sys := range gs.Data.Systems {
		if gs.Systems[i].Event != "" {
			dist := formula.Distance(cur.X, cur.Y, sys.X, sys.Y)
			if dist < 30 {
				return
			}
		}
	}

	var candidates []int
	for i, sys := range gs.Data.Systems {
		if i == gs.CurrentSystemID {
			continue
		}
		dist := formula.Distance(cur.X, cur.Y, sys.X, sys.Y)
		if dist < 30 {
			candidates = append(candidates, i)
		}
	}
	if len(candidates) == 0 {
		return
	}

	gs.Rand.Shuffle(len(candidates), func(i, j int) {
		candidates[i], candidates[j] = candidates[j], candidates[i]
	})
	for _, idx := range candidates {
		event := eventNames[gs.Rand.Intn(len(eventNames))]
		if eventHasTradeableGoods(gs, idx, event) {
			gs.Systems[idx].Event = event
			gs.Systems[idx].EventDay = gs.Day
			RefreshSystemPrices(gs, idx)
			return
		}
	}
}

func initializeMarkets(gs *GameState) {
	for i := range gs.Data.Systems {
		RefreshSystemPrices(gs, i)
	}
}

func RefreshOtherSystemPrices(gs *GameState, skipIdx int) {
	for i := range gs.Data.Systems {
		if i != skipIdx {
			RefreshSystemPrices(gs, i)
		}
	}
}

func RefreshSystemPrices(gs *GameState, sysIdx int) {
	sys := gs.Data.Systems[sysIdx]
	event := gs.Systems[sysIdx].Event
	for g, good := range gs.Data.Goods {
		if good.MinTech <= sys.TechLevel && sys.TechLevel <= good.MaxTech {
			gs.Systems[sysIdx].Prices[g] = formula.BasePrice(good, sys, event, gs.Rand)
		} else {
			gs.Systems[sysIdx].Prices[g] = -1
		}
	}
}

