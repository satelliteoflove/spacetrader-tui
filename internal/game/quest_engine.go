package game

import (
	"fmt"

	"github.com/the4ofus/spacetrader-tui/internal/formula"
	"github.com/the4ofus/spacetrader-tui/internal/gamedata"
)

type QuestEvent struct {
	Title   string
	Message string
	Actions []string
}

func CheckQuestsOnArrival(gs *GameState) []QuestEvent {
	var events []QuestEvent

	if gs.Day > 10 && gs.Quests.States[QuestMoonForSale] == QuestUnavailable {
		gs.Quests.States[QuestMoonForSale] = QuestAvailable
	}

	sys := gs.Data.Systems[gs.CurrentSystemID]

	if gs.Day > 5 && gs.Quests.States[QuestJapori] == QuestUnavailable && sys.TechLevel >= gamedata.TechIndustrial {
		if gs.Rand.Intn(100) < 15 {
			gs.Quests.States[QuestJapori] = QuestAvailable
			events = append(events, QuestEvent{
				Title:   "Japori Disease",
				Message: "A terrible disease is sweeping Japori! 10 bays of medicine are desperately needed.\n\n  Reward: Skill training (+2 random skills)\n  Requires: 10 units of medicine in cargo\n  Deadline: None -- deliver at your own pace",
				Actions: []string{"Accept mission", "Decline"},
			})
		}
	}

	if gs.Quests.States[QuestJapori] == QuestActive {
		medicine := gs.Player.Cargo[int(gamedata.GoodMedicine)]
		if medicine >= 10 {
			gs.Player.Cargo[int(gamedata.GoodMedicine)] -= 10
			gs.Quests.States[QuestJapori] = QuestComplete
			skill1 := gs.Rand.Intn(formula.NumSkills)
			skill2 := gs.Rand.Intn(formula.NumSkills)
			gs.Player.Skills[skill1]++
			gs.Player.Skills[skill2]++
			if gs.Player.Skills[skill1] > formula.SkillMax {
				gs.Player.Skills[skill1] = formula.SkillMax
			}
			if gs.Player.Skills[skill2] > formula.SkillMax {
				gs.Player.Skills[skill2] = formula.SkillMax
			}
			events = append(events, QuestEvent{
				Title:   "Japori Disease - Complete!",
				Message: fmt.Sprintf("The medicine was delivered! Your %s and %s skills improved.", formula.SkillNames[skill1], formula.SkillNames[skill2]),
			})
		}
	}

	if gs.Day > 15 && gs.Quests.States[QuestSkillIncrease] == QuestUnavailable {
		if gs.Rand.Intn(100) < 10 && sys.TechLevel >= gamedata.TechPostIndustrial {
			gs.Quests.States[QuestSkillIncrease] = QuestAvailable
			events = append(events, QuestEvent{
				Title:   "Skill Training Available",
				Message: "A renowned trainer offers to improve one of your skills.\n\n  Cost: 3,000 credits\n  Reward: +1 random skill",
				Actions: []string{"Pay 3000 for training", "Decline"},
			})
		}
	}

	if gs.Day > 8 && gs.Quests.States[QuestLotteryWinner] == QuestUnavailable {
		if gs.Rand.Intn(100) < 3 {
			gs.Quests.States[QuestLotteryWinner] = QuestComplete
			winnings := 500 + gs.Rand.Intn(1500)
			gs.Player.Credits += winnings
			events = append(events, QuestEvent{
				Title:   "Lottery Winner!",
				Message: fmt.Sprintf("You won %d credits in the local lottery!", winnings),
			})
		}
	}

	if gs.Quests.States[QuestCargoForSale] == QuestUnavailable && gs.Rand.Intn(100) < 5 {
		goodIdx := gs.Rand.Intn(NumGoods)
		for gs.Data.Goods[goodIdx].MinTech > sys.TechLevel {
			goodIdx = gs.Rand.Intn(NumGoods)
		}
		dp := &GameDataProvider{Data: gs.Data}
		free := gs.Player.FreeCargo(dp)
		qty := 3 + gs.Rand.Intn(5)
		if qty > free {
			qty = free
		}
		if qty > 0 {
			good := gs.Data.Goods[goodIdx]
			price := good.BasePrice * qty / 2
			legalNote := ""
			if !good.Legal {
				legalNote = " (illegal goods!)"
			}
			gs.Quests.States[QuestCargoForSale] = QuestAvailable
			gs.Quests.Progress[QuestCargoForSale] = goodIdx
			gs.cargoOfferQty = qty
			events = append(events, QuestEvent{
				Title:   "Cargo for Sale",
				Message: fmt.Sprintf("A merchant offers %d units of %s at half price for %d credits.%s", qty, good.Name, price, legalNote),
				Actions: []string{fmt.Sprintf("Buy %d %s for %d cr", qty, good.Name, price), "Decline"},
			})
		}
	}

	if gs.Quests.States[QuestEraseRecord] == QuestUnavailable && gs.Player.PoliceRecord < -10 {
		if gs.Rand.Intn(100) < 8 && sys.PoliticalSystem == gamedata.PolAnarchy {
			gs.Quests.States[QuestEraseRecord] = QuestAvailable
			events = append(events, QuestEvent{
				Title:   "Hacker Contact",
				Message: "A hacker offers to erase your police record for 5000 credits.",
				Actions: []string{"Pay 5000", "Decline"},
			})
		}
	}

	if gs.Quests.TribbleQty > 0 {
		gs.Quests.TribbleQty = gs.Quests.TribbleQty * 2
		if gs.Quests.TribbleQty > 100 {
			gs.Quests.TribbleQty = 100
		}
		food := gs.Player.Cargo[int(gamedata.GoodFood)]
		if food > 0 {
			eaten := food
			if eaten > gs.Quests.TribbleQty/10+1 {
				eaten = gs.Quests.TribbleQty/10 + 1
			}
			gs.Player.Cargo[int(gamedata.GoodFood)] -= eaten
			events = append(events, QuestEvent{
				Title:   "Tribbles!",
				Message: fmt.Sprintf("Your %d tribbles ate %d units of food!", gs.Quests.TribbleQty, eaten),
			})
		}
	}

	chainChecks := []func(*GameState) []QuestEvent{
		checkDragonfly,
		checkSpaceMonster,
		checkScarab,
		checkAlienArtifact,
		checkJarek,
		checkGemulon,
		checkFehler,
		checkWild,
		checkReactor,
	}
	for _, check := range chainChecks {
		if evts := check(gs); len(evts) > 0 {
			events = append(events, evts...)
		}
	}

	return events
}

func ResolveQuestAction(gs *GameState, questTitle string, actionIdx int) string {
	switch questTitle {
	case "Japori Disease":
		if actionIdx == 0 {
			gs.Quests.States[QuestJapori] = QuestActive
			return "Mission accepted. Deliver 10 medicine to any system."
		}
		gs.Quests.States[QuestJapori] = QuestUnavailable
		return "Mission declined."

	case "Skill Training Available":
		if actionIdx == 0 && gs.Player.Credits >= 3000 {
			gs.Player.Credits -= 3000
			skill := gs.Rand.Intn(formula.NumSkills)
			gs.Player.Skills[skill]++
			if gs.Player.Skills[skill] > formula.SkillMax {
				gs.Player.Skills[skill] = formula.SkillMax
			}
			gs.Quests.States[QuestSkillIncrease] = QuestComplete
			return fmt.Sprintf("Your %s skill improved!", formula.SkillNames[skill])
		} else if actionIdx == 0 {
			return "Not enough credits."
		}
		gs.Quests.States[QuestSkillIncrease] = QuestUnavailable
		return "Maybe next time."

	case "Cargo for Sale":
		if actionIdx == 0 {
			goodIdx := gs.Quests.Progress[QuestCargoForSale]
			qty := gs.cargoOfferQty
			good := gs.Data.Goods[goodIdx]
			price := good.BasePrice * qty / 2
			if gs.Player.Credits < price {
				return "Not enough credits."
			}
			gs.Player.Credits -= price
			gs.Player.Cargo[goodIdx] += qty
			gs.Quests.States[QuestCargoForSale] = QuestComplete
			return fmt.Sprintf("Bought %d %s for %d credits (half price!).", qty, good.Name, price)
		}
		gs.Quests.States[QuestCargoForSale] = QuestUnavailable
		return "Declined."

	case "Hacker Contact":
		if actionIdx == 0 && gs.Player.Credits >= 5000 {
			gs.Player.Credits -= 5000
			gs.Player.PoliceRecord = 0
			gs.Quests.States[QuestEraseRecord] = QuestComplete
			return "Your police record has been erased!"
		} else if actionIdx == 0 {
			return "Not enough credits."
		}
		gs.Quests.States[QuestEraseRecord] = QuestUnavailable
		return "Declined."
	}

	if result := resolveQuestChainAction(gs, questTitle, actionIdx); result != "" {
		return result
	}
	return ""
}
