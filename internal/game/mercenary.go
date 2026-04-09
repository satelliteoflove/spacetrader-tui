package game

import "github.com/the4ofus/spacetrader-tui/internal/formula"

type Mercenary struct {
	Name      string                 `json:"name"`
	Skills    [formula.NumSkills]int `json:"skills"`
	SystemIdx int                    `json:"system_idx"`
	IsQuest   bool                   `json:"is_quest"`
}

func (m Mercenary) GetSkills() [formula.NumSkills]int {
	return m.Skills
}

func (m Mercenary) Wage() int {
	if m.IsQuest {
		return 0
	}
	sum := 0
	for _, s := range m.Skills {
		sum += s
	}
	return sum * 3
}
