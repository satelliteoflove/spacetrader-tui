package formula

import "github.com/the4ofus/spacetrader-tui/internal/gamedata"

type Mercenary interface {
	GetSkills() [NumSkills]int
}

func EffectiveSkill(playerSkill int, crew []Mercenary, skillIdx int, gadgetBonus int, diff gamedata.Difficulty) int {
	best := playerSkill
	for _, m := range crew {
		s := m.GetSkills()[skillIdx]
		if s > best {
			best = s
		}
	}
	result := best + gadgetBonus

	switch diff {
	case gamedata.DiffBeginner, gamedata.DiffEasy:
		result++
	case gamedata.DiffImpossible:
		result--
	}

	if result < 1 {
		result = 1
	}
	return result
}
