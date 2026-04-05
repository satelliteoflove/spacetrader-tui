package screens

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/the4ofus/spacetrader-tui/internal/game"
)

type SaveScreen struct {
	gs      *game.GameState
	cursor  int
	message string
	items   []string
}

func NewSaveScreen(gs *game.GameState) *SaveScreen {
	return &SaveScreen{
		gs:    gs,
		items: []string{"Save Game", "Back"},
	}
}

func (s *SaveScreen) Init() tea.Cmd { return nil }

func (s *SaveScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, Keys.Up):
			s.cursor = wrapCursor(s.cursor, -1, len(s.items))
		case key.Matches(msg, Keys.Down):
			s.cursor = wrapCursor(s.cursor, 1, len(s.items))
		case key.Matches(msg, Keys.Enter):
			switch s.cursor {
			case 0:
				path, err := game.DefaultSavePath()
				if err != nil {
					s.message = fmt.Sprintf("Error: %v", err)
					return s, nil
				}
				if err := game.Save(s.gs, path); err != nil {
					s.message = fmt.Sprintf("Save failed: %v", err)
				} else {
					s.message = fmt.Sprintf("Game saved to %s", path)
				}
			case 1:
				return s, func() tea.Msg { return NavigateMsg{Screen: ScreenSystem} }
			}
		case key.Matches(msg, Keys.Back):
			return s, func() tea.Msg { return NavigateMsg{Screen: ScreenSystem} }
		}
	}
	return s, nil
}

func (s *SaveScreen) View() string {
	var b strings.Builder

	b.WriteString(HeaderStyle.Render("  SAVE GAME  ") + "\n\n")
	b.WriteString(fmt.Sprintf("  Commander: %s  |  Day: %d  |  Credits: %d\n\n",
		s.gs.Player.Name, s.gs.Day, s.gs.Player.Credits))

	for i, item := range s.items {
		if i == s.cursor {
			b.WriteString(fmt.Sprintf("  %s\n", SelectedStyle.Render("> "+item)))
		} else {
			b.WriteString(fmt.Sprintf("    %s\n", NormalStyle.Render(item)))
		}
	}

	if s.message != "" {
		b.WriteString("\n  " + s.message + "\n")
	}

	b.WriteString("\n" + DimStyle.Render("  enter select, esc back"))
	return b.String()
}
