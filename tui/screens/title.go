package screens

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const asciiTitle = `   _____                          ______               __
  / ___/____  ____ _________     /_  __/________ _____/ /__  _____
  \__ \/ __ \/ __ ` + "`" + `/ ___/ _ \     / / / ___/ __ ` + "`" + `/ __  / _ \/ ___/
 ___/ / /_/ / /_/ / /__/  __/    / / / /  / /_/ / /_/ /  __/ /
/____/ .___/\__,_/\___/\___/    /_/ /_/   \__,_/\__,_/\___/_/
    /_/`

type menuAction int

const (
	actionNewGame menuAction = iota
	actionLoadGame
	actionSettings
	actionQuit
)

type titleMenuItem struct {
	label  string
	action menuAction
}

type TitleScreen struct {
	cursor int
	items  []titleMenuItem
	tw     *Typewriter
}

func NewTitleScreen() *TitleScreen {
	return NewTitleScreenWithConfig(false, false)
}

func NewTitleScreenWithConfig(colorblind bool, hasSave bool) *TitleScreen {
	var items []titleMenuItem
	if hasSave {
		items = append(items, titleMenuItem{"Load Game", actionLoadGame})
	}
	items = append(items, titleMenuItem{"New Game", actionNewGame})
	items = append(items, titleMenuItem{"Settings", actionSettings})
	items = append(items, titleMenuItem{"Quit", actionQuit})

	return &TitleScreen{
		items: items,
		tw:    NewTypewriter(asciiTitle, AnimTypewriterTitle),
	}
}

func (s *TitleScreen) Init() tea.Cmd { return nil }

func (s *TitleScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case TickMsg:
		s.tw.Start(msg.Time)
		s.tw.Update(msg.Time)
	case tea.KeyMsg:
		if !s.tw.Done() {
			s.tw.Skip()
			return s, nil
		}
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
			case actionSettings:
				return s, func() tea.Msg { return NavigateMsg{Screen: ScreenSettings} }
			case actionQuit:
				return s, tea.Quit
			}
		}
	}
	return s, nil
}

func (s *TitleScreen) View() string {
	var b strings.Builder

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("14"))

	b.WriteString("\n")
	b.WriteString(titleStyle.Render(s.tw.View()))
	b.WriteString("\n")

	subtitle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		Render("Terminal Edition")

	b.WriteString(subtitle + "\n\n")

	if s.tw.Done() {
		labels := make([]string, len(s.items))
		for i, item := range s.items {
			labels[i] = item.label
		}
		RenderMenuItems(&b, labels, s.cursor)
		b.WriteString("\n" + DimStyle.Render("j/k to move, enter to select"))
	}

	return b.String()
}
