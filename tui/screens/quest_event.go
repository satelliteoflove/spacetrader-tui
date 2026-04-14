package screens

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/the4ofus/spacetrader-tui/internal/game"
)

type questEventPhase int

const (
	questPhaseMessage questEventPhase = iota
	questPhaseCombat
	questPhaseResult
)

type QuestEventScreen struct {
	gs         *game.GameState
	events     []game.QuestEvent
	current    int
	cursor     int
	phase      questEventPhase
	result     string
	combatAnim *CombatLogAnimator
	tw         *Typewriter
}

func NewQuestEventScreen(gs *game.GameState, events []game.QuestEvent) *QuestEventScreen {
	s := &QuestEventScreen{gs: gs, events: events}
	if len(events) > 0 {
		s.tw = NewTypewriter(events[0].Message, AnimTypewriterEncounter)
	}
	return s
}

func (s *QuestEventScreen) Init() tea.Cmd { return nil }

func (s *QuestEventScreen) advanceEvent() {
	s.current++
	s.cursor = 0
	s.phase = questPhaseMessage
	s.result = ""
	s.combatAnim = nil
	if s.current < len(s.events) {
		s.tw = NewTypewriter(s.events[s.current].Message, AnimTypewriterEncounter)
	}
}

func (s *QuestEventScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case TickMsg:
		if s.tw != nil {
			s.tw.Start(msg.Time)
			s.tw.Update(msg.Time)
		}
		if s.phase == questPhaseCombat && s.combatAnim != nil {
			s.combatAnim.Update(msg.Time)
			if s.combatAnim.Done() {
				s.phase = questPhaseResult
				s.tw = NewTypewriter(s.result, AnimTypewriterEncounter)
			}
		}
		return s, nil
	case tea.KeyMsg:
		if s.tw != nil && !s.tw.Done() {
			s.tw.Skip()
			return s, nil
		}

		switch s.phase {
		case questPhaseMessage:
			evt := s.events[s.current]
			if len(evt.Actions) == 0 {
				if key.Matches(msg, Keys.Enter) || key.Matches(msg, Keys.Back) {
					s.advanceEvent()
					if s.current >= len(s.events) {
						return s, func() tea.Msg { return NavigateMsg{Screen: ScreenSystem} }
					}
				}
				return s, nil
			}
			switch {
			case key.Matches(msg, Keys.Up):
				s.cursor = wrapCursor(s.cursor, -1, len(evt.Actions))
			case key.Matches(msg, Keys.Down):
				s.cursor = wrapCursor(s.cursor, 1, len(evt.Actions))
			case key.Matches(msg, Keys.Enter):
				actionResult := game.ResolveQuestAction(s.gs, evt.Title, s.cursor)
				if actionResult.Combat != nil {
					combat := actionResult.Combat
					if len(combat.Log) > 0 {
						s.phase = questPhaseCombat
						s.combatAnim = NewCombatLogAnimator(combat.Log)
						s.result = combat.Result
					} else {
						s.phase = questPhaseResult
						s.result = combat.Result
						s.tw = NewTypewriter(s.result, AnimTypewriterEncounter)
					}
				} else if actionResult.Message != "" {
					s.phase = questPhaseResult
					s.result = actionResult.Message
					s.tw = NewTypewriter(s.result, AnimTypewriterEncounter)
				} else {
					s.advanceEvent()
					if s.current >= len(s.events) {
						return s, func() tea.Msg { return NavigateMsg{Screen: ScreenSystem} }
					}
				}
			}

		case questPhaseCombat:
			if s.combatAnim != nil && !s.combatAnim.Done() {
				s.combatAnim.Skip()
			}

		case questPhaseResult:
			if key.Matches(msg, Keys.Enter) {
				s.advanceEvent()
				if s.current >= len(s.events) {
					return s, func() tea.Msg { return NavigateMsg{Screen: ScreenSystem} }
				}
			}
		}
	}
	return s, nil
}

func (s *QuestEventScreen) View() string {
	var b strings.Builder

	if s.current >= len(s.events) {
		return ""
	}

	evt := s.events[s.current]
	b.WriteString(HeaderStyle.Render(fmt.Sprintf("  %s  ", evt.Title)) + "\n\n")

	switch s.phase {
	case questPhaseMessage:
		b.WriteString("  " + s.tw.View() + "\n\n")
		twDone := s.tw == nil || s.tw.Done()
		if len(evt.Actions) > 0 && twDone {
			RenderMenuItems(&b, evt.Actions, s.cursor)
			b.WriteString("\n" + DimStyle.Render("  j/k choose, enter select"))
		} else if len(evt.Actions) == 0 && twDone {
			b.WriteString(SelectedStyle.Render("  press enter to continue"))
		}

	case questPhaseCombat:
		b.WriteString("  " + evt.Message + "\n\n")
		if s.combatAnim != nil {
			b.WriteString(s.combatAnim.View())
		}

	case questPhaseResult:
		b.WriteString("  " + evt.Message + "\n\n")
		if s.combatAnim != nil {
			b.WriteString(s.combatAnim.StaticView())
			b.WriteString("\n")
		}
		b.WriteString("  " + SelectedStyle.Render(s.tw.View()) + "\n")
		if s.tw != nil && s.tw.Done() {
			b.WriteString("\n" + SelectedStyle.Render("  press enter to continue"))
		}
	}

	return b.String()
}
