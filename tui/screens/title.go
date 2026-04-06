package screens

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

)

type TitleScreen struct {
	cursor    int
	items     []string
	colorblind bool
}

func NewTitleScreen() *TitleScreen {
	return NewTitleScreenWithConfig(false)
}

func NewTitleScreenWithConfig(colorblind bool) *TitleScreen {
	cbLabel := "Colorblind Mode: OFF"
	if colorblind {
		cbLabel = "Colorblind Mode: ON"
	}
	return &TitleScreen{
		items:     []string{"New Game", "Load Game", cbLabel, "Quit"},
		colorblind: colorblind,
	}
}

func (s *TitleScreen) Init() tea.Cmd { return nil }

func (s *TitleScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
				return s, func() tea.Msg { return NavigateMsg{Screen: ScreenNewGame} }
			case 1:
				return s, func() tea.Msg { return LoadGameMsg{} }
			case 2:
				return s, func() tea.Msg { return ToggleColorblindMsg{} }
			case 3:
				return s, tea.Quit
			}
		}
	}
	return s, nil
}

func (s *TitleScreen) View() string {
	var b strings.Builder

	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("14")).
		Padding(2, 0, 1, 0).
		Render("SPACE TRADER")

	subtitle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		Render("Terminal Edition")

	b.WriteString(title + "\n")
	b.WriteString(subtitle + "\n\n")

	RenderMenuItems(&b, s.items, s.cursor)

	b.WriteString("\n" + DimStyle.Render("j/k to move, enter to select"))

	return b.String()
}
