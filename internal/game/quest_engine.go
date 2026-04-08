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

	if gs.Day > 10 && gs.QuestState(QuestMoonForSale) == QuestUnavailable {
		gs.SetQuestState(QuestMoonForSale, QuestAvailable)
	}

	sys := gs.Data.Systems[gs.CurrentSystemID]

	if gs.Day > 5 && gs.QuestState(QuestJapori) == QuestUnavailable && sys.TechLevel >= gamedata.TechIndustrial {
		if gs.Rand.Intn(100) < 15 {
			gs.SetQuestState(QuestJapori, QuestAvailable)
			events = append(events, QuestEvent{
				Title:   "Japori Disease",
				Message: "A terrible disease is sweeping Japori! 10 bays of medicine are desperately needed.\n\n  Reward: Skill training (+2 random skills)\n  Requires: 10 units of medicine in cargo\n  Deadline: None -- deliver at your own pace",
				Actions: []string{"Accept mission", "Decline"},
			})
		}
	}

	japoriSys := findSystem(gs, "Japori")
	if gs.QuestState(QuestJapori) == QuestActive && japoriSys >= 0 && gs.CurrentSystemID == japoriSys {
		medicine := gs.Player.Cargo[int(gamedata.GoodMedicine)]
		if medicine >= 10 {
			gs.Player.Cargo[int(gamedata.GoodMedicine)] -= 10
			gs.SetQuestState(QuestJapori, QuestComplete)
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

	if gs.Day > 15 && gs.QuestState(QuestSkillIncrease) == QuestUnavailable {
		if gs.Rand.Intn(100) < 10 && sys.TechLevel >= gamedata.TechPostIndustrial {
			gs.SetQuestState(QuestSkillIncrease, QuestAvailable)
			events = append(events, QuestEvent{
				Title:   "Skill Training Available",
				Message: "A renowned trainer offers to improve one of your skills.\n\n  Cost: 3,000 credits\n  Reward: +1 random skill",
				Actions: []string{"Pay 3000 for training", "Decline"},
			})
		}
	}

	if gs.Day > 8 && gs.QuestState(QuestLotteryWinner) == QuestUnavailable {
		if gs.Rand.Intn(100) < 3 {
			gs.SetQuestState(QuestLotteryWinner, QuestComplete)
			winnings := 500 + gs.Rand.Intn(1500)
			gs.Player.Credits += winnings
			events = append(events, QuestEvent{
				Title:   "Lottery Winner!",
				Message: fmt.Sprintf("You won %d credits in the local lottery!", winnings),
			})
		}
	}

	if gs.QuestState(QuestCargoForSale) == QuestUnavailable && gs.Rand.Intn(100) < 5 {
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
			gs.SetQuestState(QuestCargoForSale, QuestAvailable)
			gs.SetQuestProgress(QuestCargoForSale, goodIdx)
			gs.cargoOfferQty = qty
			events = append(events, QuestEvent{
				Title:   "Cargo for Sale",
				Message: fmt.Sprintf("A merchant offers %d units of %s at half price for %d credits.%s", qty, good.Name, price, legalNote),
				Actions: []string{fmt.Sprintf("Buy %d %s for %d cr", qty, good.Name, price), "Decline"},
			})
		}
	}

	if gs.QuestState(QuestEraseRecord) == QuestUnavailable && gs.Player.PoliceRecord < -10 {
		if gs.Rand.Intn(100) < 8 && sys.PoliticalSystem == gamedata.PolAnarchy {
			gs.SetQuestState(QuestEraseRecord, QuestAvailable)
			events = append(events, QuestEvent{
				Title:   "Hacker Contact",
				Message: "A hacker offers to erase your police record for 5000 credits.",
				Actions: []string{"Pay 5000", "Decline"},
			})
		}
	}

	if gs.Quests.TribbleQty > 0 {
		narcIdx := int(gamedata.GoodNarcotics)
		if gs.Player.Cargo[narcIdx] > 0 {
			narcQty := gs.Player.Cargo[narcIdx]
			gs.Player.Cargo[narcIdx] = 0
			fursIdx := int(gamedata.GoodFurs)
			gs.Player.Cargo[fursIdx] += narcQty
			gs.Quests.TribbleQty = 1 + gs.Rand.Intn(3)
			events = append(events, QuestEvent{
				Title:   "Tribbles!",
				Message: fmt.Sprintf("The tribbles consumed your narcotics and mostly died! Only %d remain. You're left with %d furs.", gs.Quests.TribbleQty, narcQty),
			})
		} else {
			foodIdx := int(gamedata.GoodFood)
			food := gs.Player.Cargo[foodIdx]
			divisor := 2
			if food > 0 {
				divisor = 1
			}
			breed := 1 + gs.Rand.Intn(max(1, gs.Quests.TribbleQty/divisor))
			gs.Quests.TribbleQty += breed

			if food > 0 {
				foodGrowth := 100 + gs.Rand.Intn(food*100+1)
				gs.Quests.TribbleQty += foodGrowth
				gs.Player.Cargo[foodIdx] = 0
				events = append(events, QuestEvent{
					Title:   "Tribbles!",
					Message: fmt.Sprintf("Your %d tribbles ate all your food and multiplied wildly!", gs.Quests.TribbleQty),
				})
			}

			if gs.Quests.TribbleQty > 100000 {
				gs.Quests.TribbleQty = 100000
			}
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
			gs.SetQuestState(QuestJapori, QuestActive)
			return "Mission accepted. Deliver 10 medicine to Japori."
		}
		gs.SetQuestState(QuestJapori, QuestUnavailable)
		return "Mission declined."

	case "Skill Training Available":
		if actionIdx == 0 && gs.Player.Credits >= 3000 {
			gs.Player.Credits -= 3000
			skill := gs.Rand.Intn(formula.NumSkills)
			gs.Player.Skills[skill]++
			if gs.Player.Skills[skill] > formula.SkillMax {
				gs.Player.Skills[skill] = formula.SkillMax
			}
			gs.SetQuestState(QuestSkillIncrease, QuestComplete)
			return fmt.Sprintf("Your %s skill improved!", formula.SkillNames[skill])
		} else if actionIdx == 0 {
			return "Not enough credits."
		}
		gs.SetQuestState(QuestSkillIncrease, QuestUnavailable)
		return "Maybe next time."

	case "Cargo for Sale":
		if actionIdx == 0 {
			goodIdx := gs.QuestProgress(QuestCargoForSale)
			qty := gs.cargoOfferQty
			good := gs.Data.Goods[goodIdx]
			price := good.BasePrice * qty / 2
			if gs.Player.Credits < price {
				return "Not enough credits."
			}
			gs.Player.Credits -= price
			gs.Player.Cargo[goodIdx] += qty
			gs.SetQuestState(QuestCargoForSale, QuestComplete)
			return fmt.Sprintf("Bought %d %s for %d credits (half price!).", qty, good.Name, price)
		}
		gs.SetQuestState(QuestCargoForSale, QuestUnavailable)
		return "Declined."

	case "Hacker Contact":
		if actionIdx == 0 && gs.Player.Credits >= 5000 {
			gs.Player.Credits -= 5000
			gs.Player.PoliceRecord = 0
			gs.SetQuestState(QuestEraseRecord, QuestComplete)
			return "Your police record has been erased!"
		} else if actionIdx == 0 {
			return "Not enough credits."
		}
		gs.SetQuestState(QuestEraseRecord, QuestUnavailable)
		return "Declined."
	}

	if result := resolveQuestChainAction(gs, questTitle, actionIdx); result != "" {
		return result
	}
	return ""
}
