package screens

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type menuAction int

const (
	actionNewGame menuAction = iota
	actionLoadGame
	actionColorblind
	actionQuit
)

type titleMenuItem struct {
	label  string
	action menuAction
}

type TitleScreen struct {
	cursor     int
	items      []titleMenuItem
	colorblind bool
}

func NewTitleScreen() *TitleScreen {
	return NewTitleScreenWithConfig(false, false)
}

func NewTitleScreenWithConfig(colorblind bool, hasSave bool) *TitleScreen {
	cbLabel := "Colorblind Mode: OFF"
	if colorblind {
		cbLabel = "Colorblind Mode: ON"
	}

	var items []titleMenuItem
	if hasSave {
		items = append(items, titleMenuItem{"Load Game", actionLoadGame})
	}
	items = append(items, titleMenuItem{"New Game", actionNewGame})
	items = append(items, titleMenuItem{cbLabel, actionColorblind})
	items = append(items, titleMenuItem{"Quit", actionQuit})

	return &TitleScreen{
		items:      items,
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
			switch s.items[s.cursor].action {
			case actionNewGame:
				return s, func() tea.Msg { return NavigateMsg{Screen: ScreenNewGame} }
			case actionLoadGame:
				return s, func() tea.Msg { return LoadGameMsg{} }
			case actionColorblind:
				return s, func() tea.Msg { return ToggleColorblindMsg{} }
			case actionQuit:
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

	labels := make([]string, len(s.items))
	for i, item := range s.items {
		labels[i] = item.label
	}
	RenderMenuItems(&b, labels, s.cursor)

	b.WriteString("\n" + DimStyle.Render("j/k to move, enter to select"))

	return b.String()
}
