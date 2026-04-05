package game

import "github.com/the4ofus/spacetrader-tui/internal/formula"

type Mercenary struct {
	Name   string                 `json:"name"`
	Skills [formula.NumSkills]int `json:"skills"`
	Wage   int                    `json:"wage"`
}

func (m Mercenary) GetSkills() [formula.NumSkills]int {
	return m.Skills
}
