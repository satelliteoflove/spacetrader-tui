package travel

import (
	"fmt"
	"math"
	"sort"

	"github.com/the4ofus/spacetrader-tui/internal/formula"
	"github.com/the4ofus/spacetrader-tui/internal/game"
)

type ReachableSystem struct {
	Index    int
	Name     string
	Distance float64
}

func ReachableSystems(gs *game.GameState) []ReachableSystem {
	cur := gs.Data.Systems[gs.CurrentSystemID]
	fuel := gs.Player.Ship.Fuel

	var reachable []ReachableSystem
	for i, sys := range gs.Data.Systems {
		if i == gs.CurrentSystemID {
			continue
		}
		dist := formula.Distance(cur.X, cur.Y, sys.X, sys.Y)
		intDist := int(math.Ceil(dist))
		if intDist <= fuel {
			reachable = append(reachable, ReachableSystem{
				Index:    i,
				Name:     sys.Name,
				Distance: dist,
			})
		}
	}
	sort.Slice(reachable, func(i, j int) bool {
		return reachable[i].Distance < reachable[j].Distance
	})
	return reachable
}

type TravelResult struct {
	Success     bool
	Message     string
	FuelUsed    int
	DaysElapsed int
}

func ExecuteTravel(gs *game.GameState, destIdx int) TravelResult {
	if destIdx < 0 || destIdx >= len(gs.Data.Systems) {
		return TravelResult{Message: "Invalid destination."}
	}
	if destIdx == gs.CurrentSystemID {
		return TravelResult{Message: "Already at this system."}
	}

	cur := gs.Data.Systems[gs.CurrentSystemID]
	dest := gs.Data.Systems[destIdx]
	dist := formula.Distance(cur.X, cur.Y, dest.X, dest.Y)
	fuelNeeded := int(math.Ceil(dist))

	if fuelNeeded > gs.Player.Ship.Fuel {
		return TravelResult{Message: "Not enough fuel."}
	}

	gs.Player.Ship.Fuel -= fuelNeeded
	gs.CurrentSystemID = destIdx
	gs.Systems[destIdx].Visited = true
	gs.Day++

	applyDailyCosts(gs)
	applyEngineerRepair(gs)
	game.GenerateEvents(gs)
	game.RefreshSystemPrices(gs, destIdx)

	return TravelResult{
		Success:     true,
		Message:     fmt.Sprintf("Arrived at %s. Day %d.", dest.Name, gs.Day),
		FuelUsed:    fuelNeeded,
		DaysElapsed: 1,
	}
}

func applyDailyCosts(gs *game.GameState) {
	totalWages := 0
	for _, m := range gs.Player.Crew {
		totalWages += m.Wage
	}
	gs.Player.Credits -= totalWages
	if gs.Player.Credits < 0 {
		gs.Player.Credits = 0
		gs.Player.Crew = nil
	}

	if gs.Player.LoanBalance > 0 {
		interest := gs.Player.LoanBalance / 10
		if interest < 1 {
			interest = 1
		}
		gs.Player.LoanBalance += interest
	}

	if gs.Player.HasInsurance {
		gs.Player.InsuranceDays++
		basePremium := 100
		noclaimDiscount := gs.Player.InsuranceDays
		if noclaimDiscount > 90 {
			noclaimDiscount = 90
		}
		premium := basePremium * (100 - noclaimDiscount) / 100
		if premium < 10 {
			premium = 10
		}
		gs.Player.Credits -= premium
		if gs.Player.Credits < 0 {
			gs.Player.Credits = 0
			gs.Player.HasInsurance = false
		}
	}
}

func applyEngineerRepair(gs *game.GameState) {
	engSkill := formula.EffectiveSkill(gs.Player.Skills[formula.SkillEngineer], gs.Player.CrewMercs(), formula.SkillEngineer, 0)

	shipDef := gs.Data.Ships[gs.Player.Ship.TypeID]
	if gs.Player.Ship.Hull < shipDef.Hull {
		repair := engSkill / 2
		if repair < 1 {
			repair = 1
		}
		gs.Player.Ship.Hull += repair
		if gs.Player.Ship.Hull > shipDef.Hull {
			gs.Player.Ship.Hull = shipDef.Hull
		}
	}
}
