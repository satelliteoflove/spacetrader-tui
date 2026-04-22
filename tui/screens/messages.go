package screens

import (
	"time"

	"github.com/the4ofus/spacetrader-tui/internal/encounter"
	"github.com/the4ofus/spacetrader-tui/internal/formula"
	"github.com/the4ofus/spacetrader-tui/internal/gamedata"
)

type StartGameMsg struct {
	Name       string
	Skills     [formula.NumSkills]int
	Difficulty gamedata.Difficulty
}

type TravelMsg struct {
	DestIdx int
}

type EncounterDoneMsg struct {
	Outcome encounter.Outcome
}

type LoadGameMsg struct {
	FromAutosave bool
}

type TickMsg struct {
	Time time.Time
}

type WarpDoneMsg struct{}

type WarpEncounterMsg struct {
	Encounter *encounter.Encounter
}

type WarpResumeMsg struct{}
