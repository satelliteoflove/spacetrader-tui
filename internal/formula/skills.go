package formula

type Mercenary interface {
	GetSkills() [NumSkills]int
}

func EffectiveSkill(playerSkill int, crew []Mercenary, skillIdx int, gadgetBonus int) int {
	best := playerSkill
	for _, m := range crew {
		s := m.GetSkills()[skillIdx]
		if s > best {
			best = s
		}
	}
	return best + gadgetBonus
}
