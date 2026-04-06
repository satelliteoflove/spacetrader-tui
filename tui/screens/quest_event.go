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
}

func NewQuestEventScreen(gs *game.GameState, events []game.QuestEvent) *QuestEventScreen {
	return &QuestEventScreen{gs: gs, events: events}
}

func (s *QuestEventScreen) Init() tea.Cmd { return nil }

func (s *QuestEventScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if s.result != "" {
			if key.Matches(msg, Keys.Enter) {
				s.result = ""
				s.current++
				if s.current >= len(s.events) {
					return s, func() tea.Msg { return NavigateMsg{Screen: ScreenSystem} }
				}
				s.cursor = 0
				return s, nil
			}
			return s, nil
		}

		evt := s.events[s.current]
		if len(evt.Actions) == 0 {
			if key.Matches(msg, Keys.Enter) || key.Matches(msg, Keys.Back) {
				s.current++
				if s.current >= len(s.events) {
					return s, func() tea.Msg { return NavigateMsg{Screen: ScreenSystem} }
				}
				s.cursor = 0
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
				s.current++
				if s.current >= len(s.events) {
					return s, func() tea.Msg { return NavigateMsg{Screen: ScreenSystem} }
				}
				s.cursor = 0
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
	b.WriteString("  " + evt.Message + "\n\n")

	if s.result != "" {
		b.WriteString("  " + SuccessStyle.Render(s.result) + "\n")
		b.WriteString("\n" + DimStyle.Render("  press enter to continue"))
	} else if len(evt.Actions) > 0 {
		RenderMenuItems(&b, evt.Actions, s.cursor)
		b.WriteString("\n" + DimStyle.Render("  j/k choose, enter select"))
	} else {
		b.WriteString(DimStyle.Render("  press enter to continue"))
	}

	return b.String()
}
