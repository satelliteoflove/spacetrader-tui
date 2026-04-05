package screens

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"
)

type ScreenType int

const (
	ScreenTitle ScreenType = iota
	ScreenNewGame
	ScreenSystem
	ScreenMarket
	ScreenChart
	ScreenEncounter
	ScreenShipyard
	ScreenBank
	ScreenStatus
	ScreenPersonnel
	ScreenGalacticChart
	ScreenGalacticList
	ScreenSave
	ScreenGameOver
)

type NavigateMsg struct {
	Screen     ScreenType
	RestoreCursor int
}

func wrapCursor(cursor, delta, length int) int {
	if length == 0 {
		return 0
	}
	return (cursor + delta + length) % length
}

type KeyMap struct {
	Up    key.Binding
	Down  key.Binding
	Enter key.Binding
	Back  key.Binding
	Quit  key.Binding
}

var Keys = KeyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
	),
	Back: key.NewBinding(
		key.WithKeys("esc", "q"),
	),
	Quit: key.NewBinding(
		key.WithKeys("ctrl+c"),
	),
}

var (
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("15")).
			Padding(1, 0)

	SelectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("11")).
			Bold(true)

	NormalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("7"))

	DangerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("9")).
			Bold(true)

	SuccessStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("10"))

	DimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("8"))

	HeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("14")).
			BorderBottom(true).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("8"))

	IllegalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("9"))
)
