package screens

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/the4ofus/spacetrader-tui/internal/encounter"
	"github.com/the4ofus/spacetrader-tui/internal/game"
)

type encounterPhase int

const (
	phaseChoose encounterPhase = iota
	phaseCombat
	phaseResult
)

type EncounterScreen struct {
	gs            *game.GameState
	enc           *encounter.Encounter
	phase         encounterPhase
	cursor        int
	outcome       encounter.Outcome
	combatAnim    *CombatLogAnimator
	entranceDelay int
	tw            *Typewriter
}

func NewEncounterScreen(gs *game.GameState, enc *encounter.Encounter) *EncounterScreen {
	return &EncounterScreen{
		gs:  gs,
		enc: enc,
		tw:  NewTypewriter(enc.Message, AnimTypewriterEncounter),
	}
}

func (s *EncounterScreen) Init() tea.Cmd { return nil }

func (s *EncounterScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case TickMsg:
		if s.entranceDelay < AnimEntranceThreshold {
			s.entranceDelay++
		} else {
			s.tw.Start(msg.Time)
			s.tw.Update(msg.Time)
			if s.phase == phaseCombat && s.combatAnim != nil {
				s.combatAnim.Update(msg.Time)
				if s.combatAnim.Done() {
					s.phase = phaseResult
					s.tw = NewTypewriter(s.outcome.Message, AnimTypewriterEncounter)
				}
			}
		}
		return s, nil
	case tea.KeyMsg:
		if s.entranceDelay < AnimEntranceThreshold {
			return s, nil
		}
		if !s.tw.Done() {
			s.tw.Skip()
			return s, nil
		}

		switch s.phase {
		case phaseChoose:
			switch {
			case key.Matches(msg, Keys.Up):
				s.cursor = wrapCursor(s.cursor, -1, len(s.enc.Actions))
			case key.Matches(msg, Keys.Down):
				s.cursor = wrapCursor(s.cursor, 1, len(s.enc.Actions))
			case key.Matches(msg, Keys.Enter):
				action := s.enc.Actions[s.cursor]
				s.outcome = encounter.Resolve(s.gs, s.enc, action)
				if len(s.outcome.CombatLog) > 0 {
					s.phase = phaseCombat
					logLines := s.outcome.CombatLog
					if statsLines := BuildCombatStatsLines(s.gs, s.enc.PirateShip); len(statsLines) > 0 {
						logLines = append(statsLines, logLines...)
					}
					s.combatAnim = NewCombatLogAnimator(logLines)
				} else {
					s.phase = phaseResult
					s.tw = NewTypewriter(s.outcome.Message, AnimTypewriterEncounter)
				}
			}

		case phaseCombat:
			if s.combatAnim != nil && !s.combatAnim.Done() {
				s.combatAnim.Skip()
			}

		case phaseResult:
			if !s.tw.Done() {
				s.tw.Skip()
				return s, nil
			}
			if key.Matches(msg, Keys.Enter) || key.Matches(msg, Keys.Back) {
				return s, func() tea.Msg { return EncounterDoneMsg{Outcome: s.outcome} }
			}
		}
	}
	return s, nil
}

func (s *EncounterScreen) View() string {
	if s.entranceDelay < AnimEntranceThreshold {
		return ""
	}

	var b strings.Builder

	style := TitleStyle
	if s.enc.Type == encounter.EncPirate {
		style = DangerStyle.Bold(true).Padding(1, 0)
	}
	b.WriteString(style.Render(fmt.Sprintf("ENCOUNTER: %s", s.enc.Type)) + "\n")

	switch s.phase {
	case phaseChoose:
		b.WriteString("  " + s.tw.View() + "\n")
		if s.enc.ThreatNote != "" && s.tw.Done() {
			b.WriteString("  " + DimStyle.Render(s.enc.ThreatNote) + "\n")
		}
		b.WriteString("\n")
		if s.tw.Done() {
			actionLabels := make([]string, len(s.enc.Actions))
			for i, a := range s.enc.Actions {
				label := a.String()
				if a == encounter.ActionBribe && s.enc.Type == encounter.EncPolice {
					cost := encounter.BribeCost(s.gs)
					if cost < 0 {
						label += " (impossible)"
					} else if cost > s.gs.Player.Credits {
						label += fmt.Sprintf(" (%d cr -- can't afford)", cost)
					} else {
						label += fmt.Sprintf(" (%d cr)", cost)
					}
				}
				actionLabels[i] = label
			}
			RenderMenuItems(&b, actionLabels, s.cursor)
			b.WriteString("\n" + DimStyle.Render("  j/k to choose, enter to act"))
		}

	case phaseCombat:
		b.WriteString("  " + s.enc.Message + "\n\n")
		if s.combatAnim != nil {
			b.WriteString(s.combatAnim.View())
		}

	case phaseResult:
		b.WriteString("  " + s.enc.Message + "\n\n")
		if s.combatAnim != nil {
			b.WriteString(s.combatAnim.StaticView())
			b.WriteString("\n")
		}
		b.WriteString("  " + SelectedStyle.Render(s.tw.View()) + "\n")

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
			if s.outcome.CargoGained != nil {
				for idx, qty := range s.outcome.CargoGained {
					name := s.gs.Data.Goods[idx].Name
					b.WriteString(SuccessStyle.Render(fmt.Sprintf("  Scooped: %d %s", qty, name)) + "\n")
				}
			}

			if s.gs.EndStatus == game.StatusDead {
				b.WriteString("\n" + DangerStyle.Render("  YOUR SHIP HAS BEEN DESTROYED") + "\n")
			}

			b.WriteString("\n" + SelectedStyle.Render("  press enter to continue"))
		}
	}

	return b.String()
}
