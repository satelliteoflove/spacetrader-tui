package screens

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"
)

type ScreenType int

const (
	ScreenTitle ScreenType = iota
	ScreenNewGame
	ScreenSystem
	ScreenMarket
	ScreenEncounter
	ScreenShipyard
	ScreenBank
	ScreenStatus
	ScreenPersonnel
	ScreenGalacticChart
	ScreenGalacticList
	ScreenSave
	ScreenGameOver
	ScreenGuide
	ScreenNews
)

type NavigateMsg struct {
	Screen         ScreenType
	RestoreCursor  int
	SelectedSystem int
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
	TitleStyle    lipgloss.Style
	SelectedStyle lipgloss.Style
	NormalStyle   lipgloss.Style
	DangerStyle   lipgloss.Style
	SuccessStyle  lipgloss.Style
	DimStyle      lipgloss.Style
	HeaderStyle   lipgloss.Style
	IllegalStyle  lipgloss.Style
	CyanStyle     lipgloss.Style
	MagentaStyle  lipgloss.Style
)

func init() {
	InitStyles(false)
}

func InitStyles(colorblind bool) {
	dangerColor := lipgloss.Color("9")
	successColor := lipgloss.Color("10")
	wormholeColor := lipgloss.Color("13")
	if colorblind {
		dangerColor = lipgloss.Color("208")
		successColor = lipgloss.Color("14")
		wormholeColor = lipgloss.Color("5")
	}

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
		Foreground(dangerColor).
		Bold(true)

	SuccessStyle = lipgloss.NewStyle().
		Foreground(successColor)

	DimStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("8"))

	HeaderStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("14")).
		BorderBottom(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("8"))

	IllegalStyle = lipgloss.NewStyle().
		Foreground(dangerColor)

	CyanStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("14"))

	MagentaStyle = lipgloss.NewStyle().
		Foreground(wormholeColor)
}

func WordWrap(text string, width int) string {
	words := strings.Fields(text)
	if len(words) == 0 {
		return ""
	}
	var lines []string
	line := words[0]
	for _, w := range words[1:] {
		if len(line)+1+len(w) > width {
			lines = append(lines, line)
			line = w
		} else {
			line += " " + w
		}
	}
	lines = append(lines, line)
	return strings.Join(lines, "\n")
}

func RenderMenuItems(b *strings.Builder, items []string, cursor int) {
	for i, item := range items {
		if i == cursor {
			b.WriteString("  " + SelectedStyle.Render("> "+item) + "\n")
		} else {
			b.WriteString("    " + NormalStyle.Render(item) + "\n")
		}
	}
}
