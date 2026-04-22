package travel

import (
	"fmt"
	"math"
	"sort"

	"github.com/the4ofus/spacetrader-tui/internal/economy"
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

var AutosaveEnabled bool

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

	if AutosaveEnabled {
		_ = game.Autosave(gs)
	}

	gs.Player.Ship.Fuel -= fuelNeeded
	gs.CurrentSystemID = destIdx
	gs.Systems[destIdx].Visited = true
	gs.Day++

	applyDailyCosts(gs)
	applyEngineerRepair(gs)
	applyPoliceRecordDecay(gs)
	applyQuestDailyTick(gs)

	actualDest := gs.CurrentSystemID
	actualName := gs.Data.Systems[actualDest].Name
	warped := actualDest != destIdx

	game.GenerateEvents(gs)
	game.RefreshOtherSystemPrices(gs, actualDest)
	gs.CaptureTradeInfo(actualDest)
	gs.RecordSnapshot()

	msg := fmt.Sprintf("Arrived at %s. Day %d.", actualName, gs.Day)
	if warped {
		msg = fmt.Sprintf("A tear in the timespace fabric warped you to %s! Day %d.", actualName, gs.Day)
	}

	return TravelResult{
		Success:     true,
		Message:     msg,
		FuelUsed:    fuelNeeded,
		DaysElapsed: 1,
	}
}

func ExecuteJump(gs *game.GameState, destIdx int) TravelResult {
	if destIdx < 0 || destIdx >= len(gs.Data.Systems) {
		return TravelResult{Message: "Invalid destination."}
	}
	if destIdx == gs.CurrentSystemID {
		return TravelResult{Message: "Already at this system."}
	}
	if !gs.Quests.HasSingularity {
		return TravelResult{Message: "No Portable Singularity available."}
	}

	gs.Quests.HasSingularity = false
	gs.CurrentSystemID = destIdx
	gs.Systems[destIdx].Visited = true

	game.GenerateEvents(gs)
	game.RefreshOtherSystemPrices(gs, destIdx)
	gs.CaptureTradeInfo(destIdx)
	gs.RecordSnapshot()

	dest := gs.Data.Systems[destIdx]
	return TravelResult{
		Success: true,
		Message: fmt.Sprintf("The Portable Singularity tears open space! You arrive at %s instantly.", dest.Name),
	}
}

func applyDailyCosts(gs *game.GameState) {
	applyCrewWages(gs)
	applyLoanInterest(gs)
	applyInsurancePremium(gs)
}

func applyCrewWages(gs *game.GameState) {
	totalWages := 0
	for _, m := range gs.Player.Crew {
		totalWages += m.Wage()
	}
	gs.Player.Credits -= totalWages
	if gs.Player.Credits < 0 {
		gs.Player.Credits = 0
		game.ClearCrewAndResetQuests(gs)
	}
}

func applyLoanInterest(gs *game.GameState) {
	if gs.Player.LoanBalance > 0 {
		interest := economy.LoanInterest(gs.Player.LoanBalance)
		gs.Player.LoanBalance += interest
	}
}

func applyInsurancePremium(gs *game.GameState) {
	if !gs.Player.HasInsurance {
		return
	}
	gs.Player.InsuranceDays++
	premium := game.InsuranceDailyPremium(gs)
	gs.Player.Credits -= premium
	if gs.Player.Credits < 0 {
		gs.Player.Credits = 0
		gs.Player.HasInsurance = false
		gs.Player.InsuranceDays = 0
	}
}

func applyQuestDailyTick(gs *game.GameState) {
	if gs.Quests.States[game.QuestSpaceMonster] == game.QuestAvailable ||
		gs.Quests.States[game.QuestSpaceMonster] == game.QuestActive {
		if gs.Quests.MonsterHull == 0 {
			gs.Quests.MonsterHull = game.MonsterMaxHull
		}
		if gs.Quests.MonsterHull < game.MonsterMaxHull {
			gs.Quests.MonsterHull = gs.Quests.MonsterHull * 105 / 100
			if gs.Quests.MonsterHull > game.MonsterMaxHull {
				gs.Quests.MonsterHull = game.MonsterMaxHull
			}
		}
	}

	if game.ReactorOnBoard(gs) {
		status := gs.QuestProgress(game.QuestReactor)
		status++
		gs.SetQuestProgress(game.QuestReactor, status)

		if gs.Quests.TribbleQty > 0 {
			if gs.Quests.TribbleQty < 20 {
				gs.Quests.TribbleQty = 0
			} else {
				gs.Quests.TribbleQty /= 2
			}
		}
	}

	if gs.Quests.FabricRipDays > 0 {
		isFirst := gs.Quests.FabricRipDays == 25
		gs.Quests.FabricRipDays--
		if gs.Quests.FabricRipDays <= 0 {
			gs.Quests.States[game.QuestFabricRip] = game.QuestComplete
		} else if isFirst || gs.Rand.Intn(100) < gs.Quests.FabricRipDays {
			dest := gs.Rand.Intn(len(gs.Data.Systems))
			gs.CurrentSystemID = dest
			gs.Systems[dest].Visited = true
		}
	}
}

func applyPoliceRecordDecay(gs *game.GameState) {
	record := gs.Player.PoliceRecord
	if record >= 0 {
		return
	}

	attackThresholds := [5]int{-999, -100, -70, -30, -10}
	diff := int(gs.Difficulty)
	if diff > 4 {
		diff = 4
	}
	if record < attackThresholds[diff] {
		return
	}

	gs.Player.PoliceRecord++
}

func applyEngineerRepair(gs *game.GameState) {
	engSkill := game.EffectivePlayerSkill(gs, formula.SkillEngineer)

	shipDef := gs.Data.Ships[gs.Player.Ship.TypeID]
	maxHull := shipDef.Hull
	if gs.Player.Ship.HullUpgraded {
		maxHull += game.ScarabHullBonus
	}
	if gs.Player.Ship.Hull < maxHull {
		repair := engSkill / 2
		if repair < 1 {
			repair = 1
		}
		gs.Player.Ship.Hull += repair
		if gs.Player.Ship.Hull > maxHull {
			gs.Player.Ship.Hull = maxHull
		}
	}
}
