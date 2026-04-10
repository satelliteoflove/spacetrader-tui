package game

import (
	"strings"

	"github.com/the4ofus/spacetrader-tui/internal/formula"
)

func GadgetSkillBonus(gs *GameState, skillIdx int) int {
	skillName := strings.ToLower(formula.SkillNames[skillIdx])
	bonus := 0

	for _, gID := range gs.Player.Ship.Gadgets {
		equip := gs.Data.Equipment[gID]
		if equip.SkillBonus == "" {
			continue
		}
		if equip.Name == "Cloaking Device" {
			if skillName == "pilot" {
				bonus += 2
			}
		} else if equip.SkillBonus == skillName {
			bonus += 3
		}
	}
	return bonus
}

func SellPriceAt(gs *GameState, sysIdx int, goodIdx int) int {
	basePrice := gs.Systems[sysIdx].Prices[goodIdx]
	if basePrice < 0 {
		return -1
	}
	if gs.Player.PoliceRecord < -5 {
		basePrice = basePrice * 90 / 100
	}
	if basePrice < 1 {
		basePrice = 1
	}
	return basePrice
}

func BuyPriceAt(gs *GameState, sysIdx int, goodIdx int) int {
	sellPrice := SellPriceAt(gs, sysIdx, goodIdx)
	if sellPrice < 0 {
		return -1
	}
	base := sellPrice
	if gs.Player.PoliceRecord < -5 {
		base = base * 100 / 90
	}
	traderSkill := EffectivePlayerSkill(gs, formula.SkillTrader)
	if traderSkill > 10 {
		traderSkill = 10
	}
	buyPrice := base * (103 + (10 - traderSkill)) / 100
	if buyPrice <= sellPrice {
		buyPrice = sellPrice + 1
	}
	return buyPrice
}

func EffectivePlayerSkill(gs *GameState, skillIdx int) int {
	crew := make([]formula.Mercenary, len(gs.Player.Crew))
	for i := range gs.Player.Crew {
		crew[i] = &gs.Player.Crew[i]
	}
	gadgetBonus := GadgetSkillBonus(gs, skillIdx)
	if skillIdx == formula.SkillTrader && gs.Quests.States[QuestJarek] == QuestComplete {
		gadgetBonus++
	}
	return formula.EffectiveSkill(
		gs.Player.Skills[skillIdx],
		crew,
		skillIdx,
		gadgetBonus,
		gs.Difficulty,
	)
}
