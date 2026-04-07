package screens

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/the4ofus/spacetrader-tui/internal/game"
)

type QuestEventScreen struct {
	gs      *game.GameState
	events  []game.QuestEvent
	current int
	cursor  int
	result  string
	tw      *Typewriter
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
		return s, nil
	case tea.KeyMsg:
		if s.tw != nil && !s.tw.Done() {
			s.tw.Skip()
			return s, nil
		}

		if s.result != "" {
			if key.Matches(msg, Keys.Enter) {
				s.result = ""
				s.advanceEvent()
				if s.current >= len(s.events) {
					return s, func() tea.Msg { return NavigateMsg{Screen: ScreenSystem} }
				}
				return s, nil
			}
			return s, nil
		}

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
			s.result = game.ResolveQuestAction(s.gs, evt.Title, s.cursor)
			if s.result == "" {
				s.advanceEvent()
				if s.current >= len(s.events) {
					return s, func() tea.Msg { return NavigateMsg{Screen: ScreenSystem} }
				}
			} else {
				s.tw = NewTypewriter(s.result, AnimTypewriterEncounter)
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
	if s.tw != nil && s.result == "" {
		b.WriteString("  " + s.tw.View() + "\n\n")
	} else if s.result == "" {
		b.WriteString("  " + evt.Message + "\n\n")
	} else {
		b.WriteString("  " + evt.Message + "\n\n")
	}

	twDone := s.tw == nil || s.tw.Done()

	if s.result != "" {
		b.WriteString("  " + SuccessStyle.Render(s.tw.View()) + "\n")
		if twDone {
			b.WriteString("\n" + DimStyle.Render("  press enter to continue"))
		}
	} else if len(evt.Actions) > 0 && twDone {
		RenderMenuItems(&b, evt.Actions, s.cursor)
		b.WriteString("\n" + DimStyle.Render("  j/k choose, enter select"))
	} else if len(evt.Actions) == 0 && twDone {
		b.WriteString(DimStyle.Render("  press enter to continue"))
	}

	return b.String()
}
