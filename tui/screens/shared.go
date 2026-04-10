package screens

import (
	"regexp"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"

	"github.com/the4ofus/spacetrader-tui/internal/game"
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
	ScreenSettings
	ScreenRoutePlanner
	ScreenDebug
)

type NavigateMsg struct {
	Screen         ScreenType
	RestoreCursor  int
	SelectedSystem int
	ReturnScreen   ScreenType
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
	ApplyAnimationSettings(game.Config{
		TransitionSpeed:   game.AnimMedium,
		WarpSpeed:         game.AnimMedium,
		EncounterEntrance: game.AnimMedium,
		TypewriterSpeed:   game.AnimMedium,
		PulseSpeed:        game.AnimMedium,
	})
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

var ansiRe = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)

func StripANSI(s string) string {
	return ansiRe.ReplaceAllString(s, "")
}

var FadeStyles []lipgloss.Style

var (
	AnimFadeDone            int
	AnimWarpMaxFrames       int
	AnimEntranceThreshold   int
	AnimTypewriterTitle     time.Duration
	AnimTypewriterEncounter time.Duration
	AnimPulsePhases         int
)

var fadeStepColors = []string{"232", "236", "240", "248"}

func initFadeStyles() {
	if AnimFadeDone <= 0 {
		FadeStyles = nil
		return
	}
	FadeStyles = make([]lipgloss.Style, AnimFadeDone)
	step := len(fadeStepColors) / AnimFadeDone
	if step < 1 {
		step = 1
	}
	for i := 0; i < AnimFadeDone; i++ {
		idx := i * step
		if idx >= len(fadeStepColors) {
			idx = len(fadeStepColors) - 1
		}
		FadeStyles[i] = lipgloss.NewStyle().Foreground(lipgloss.Color(fadeStepColors[idx]))
	}
}

var transitionLookup = map[game.AnimSpeed]int{
	game.AnimOff: 0, game.AnimSlow: 4, game.AnimMedium: 2, game.AnimFast: 1,
}
var warpLookup = map[game.AnimSpeed]int{
	game.AnimOff: 0, game.AnimSlow: 20, game.AnimMedium: 12, game.AnimFast: 6,
}
var entranceLookup = map[game.AnimSpeed]int{
	game.AnimOff: 0, game.AnimSlow: 4, game.AnimMedium: 2, game.AnimFast: 1,
}
var twTitleLookup = map[game.AnimSpeed]time.Duration{
	game.AnimOff: 0, game.AnimSlow: 25 * time.Millisecond, game.AnimMedium: 13 * time.Millisecond, game.AnimFast: 5 * time.Millisecond,
}
var twEncounterLookup = map[game.AnimSpeed]time.Duration{
	game.AnimOff: 0, game.AnimSlow: 70 * time.Millisecond, game.AnimMedium: 40 * time.Millisecond, game.AnimFast: 20 * time.Millisecond,
}
var pulseLookup = map[game.AnimSpeed]int{
	game.AnimOff: 0, game.AnimSlow: 16, game.AnimMedium: 12, game.AnimFast: 8,
}

func ApplyAnimationSettings(cfg game.Config) {
	AnimFadeDone = transitionLookup[cfg.TransitionSpeed]
	AnimWarpMaxFrames = warpLookup[cfg.WarpSpeed]
	AnimEntranceThreshold = entranceLookup[cfg.EncounterEntrance]
	AnimTypewriterTitle = twTitleLookup[cfg.TypewriterSpeed]
	AnimTypewriterEncounter = twEncounterLookup[cfg.TypewriterSpeed]
	AnimPulsePhases = pulseLookup[cfg.PulseSpeed]
	initFadeStyles()
}

type UpdateSettingsMsg struct {
	Config game.Config
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
