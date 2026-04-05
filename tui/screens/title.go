package screens

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

)

type TitleScreen struct {
	cursor int
	items  []string
}

func NewTitleScreen() *TitleScreen {
	return &TitleScreen{
		items: []string{"New Game", "Load Game", "Quit"},
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

	for i, item := range s.items {
		if i == s.cursor {
			b.WriteString(fmt.Sprintf("  %s\n", SelectedStyle.Render("> "+item)))
		} else {
			b.WriteString(fmt.Sprintf("    %s\n", NormalStyle.Render(item)))
		}
	}

	b.WriteString("\n" + DimStyle.Render("j/k to move, enter to select"))

	return b.String()
}
