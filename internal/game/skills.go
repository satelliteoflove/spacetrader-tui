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

func EffectivePlayerSkill(gs *GameState, skillIdx int) int {
	crew := make([]formula.Mercenary, len(gs.Player.Crew))
	for i := range gs.Player.Crew {
		crew[i] = &gs.Player.Crew[i]
	}
	gadgetBonus := GadgetSkillBonus(gs, skillIdx)
	return formula.EffectiveSkill(
		gs.Player.Skills[skillIdx],
		crew,
		skillIdx,
		gadgetBonus,
		gs.Difficulty,
	)
}
