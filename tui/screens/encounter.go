package screens

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/the4ofus/spacetrader-tui/internal/encounter"
	"github.com/the4ofus/spacetrader-tui/internal/game"
)

type encounterPhase int

const (
	phaseChoose encounterPhase = iota
	phaseResult
)

const entranceThreshold = 2

type EncounterScreen struct {
	gs            *game.GameState
	enc           *encounter.Encounter
	phase         encounterPhase
	cursor        int
	outcome       encounter.Outcome
	entranceDelay int
	tw            *Typewriter
}

func NewEncounterScreen(gs *game.GameState, enc *encounter.Encounter) *EncounterScreen {
	return &EncounterScreen{
		gs:  gs,
		enc: enc,
		tw:  NewTypewriter(enc.Message, 40*time.Millisecond),
	}
}

func (s *EncounterScreen) Init() tea.Cmd { return nil }

func (s *EncounterScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case TickMsg:
		if s.entranceDelay < entranceThreshold {
			s.entranceDelay++
		} else {
			s.tw.Start(msg.Time)
			s.tw.Update(msg.Time)
		}
		return s, nil
	case tea.KeyMsg:
		if s.entranceDelay < entranceThreshold {
			return s, nil
		}
		if !s.tw.Done() {
			s.tw.Skip()
			return s, nil
		}
		if s.phase == phaseResult {
			if key.Matches(msg, Keys.Enter) || key.Matches(msg, Keys.Back) {
				return s, func() tea.Msg { return EncounterDoneMsg{Outcome: s.outcome} }
			}
			return s, nil
		}

		switch {
		case key.Matches(msg, Keys.Up):
			s.cursor = wrapCursor(s.cursor, -1, len(s.enc.Actions))
		case key.Matches(msg, Keys.Down):
			s.cursor = wrapCursor(s.cursor, 1, len(s.enc.Actions))
		case key.Matches(msg, Keys.Enter):
			action := s.enc.Actions[s.cursor]
			s.outcome = encounter.Resolve(s.gs, s.enc, action)
			s.phase = phaseResult
			s.tw = NewTypewriter(s.outcome.Message, 40*time.Millisecond)
		}
	}
	return s, nil
}

func (s *EncounterScreen) View() string {
	if s.entranceDelay < entranceThreshold {
		return ""
	}

	var b strings.Builder

	style := TitleStyle
	if s.enc.Type == encounter.EncPirate {
		style = DangerStyle.Bold(true).Padding(1, 0)
	}
	b.WriteString(style.Render(fmt.Sprintf("ENCOUNTER: %s", s.enc.Type)) + "\n")
	b.WriteString("  " + s.tw.View() + "\n")
	if s.enc.ThreatNote != "" && s.phase == phaseChoose && s.tw.Done() {
		b.WriteString("  " + DimStyle.Render(s.enc.ThreatNote) + "\n")
	}
	b.WriteString("\n")

	if s.phase == phaseChoose {
		if s.tw.Done() {
			actionLabels := make([]string, len(s.enc.Actions))
			for i, a := range s.enc.Actions {
				actionLabels[i] = a.String()
			}
			RenderMenuItems(&b, actionLabels, s.cursor)
			b.WriteString("\n" + DimStyle.Render("  j/k to choose, enter to act"))
		}
	} else {
		b.WriteString("  " + s.tw.View() + "\n")

		if s.tw.Done() {
			if s.outcome.CreditsChange != 0 {
				if s.outcome.CreditsChange > 0 {
					b.WriteString(SuccessStyle.Render(fmt.Sprintf("  Credits: +%d", s.outcome.CreditsChange)) + "\n")
				} else {
					b.WriteString(DangerStyle.Render(fmt.Sprintf("  Credits: %d", s.outcome.CreditsChange)) + "\n")
				}
			}
			if s.outcome.HullDamage > 0 {
				b.WriteString(DangerStyle.Render(fmt.Sprintf("  Hull damage: %d", s.outcome.HullDamage)) + "\n")
			}

			if s.gs.EndStatus == game.StatusDead {
				b.WriteString("\n" + DangerStyle.Render("  YOUR SHIP HAS BEEN DESTROYED") + "\n")
			}

			b.WriteString("\n" + DimStyle.Render("  press enter to continue"))
		}
	}

	return b.String()
}
