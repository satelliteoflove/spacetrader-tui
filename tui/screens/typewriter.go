package screens

import (
	"time"

	"github.com/charmbracelet/lipgloss"
)

var fadeColors = []lipgloss.TerminalColor{
	lipgloss.AdaptiveColor{Light: "252", Dark: "236"},
	lipgloss.AdaptiveColor{Light: "249", Dark: "240"},
	lipgloss.AdaptiveColor{Light: "245", Dark: "244"},
}

type Typewriter struct {
	fullText string
	charRate time.Duration
	start    time.Time
	revealed int
	started  bool
	skipped  bool
}

func NewTypewriter(text string, charRate time.Duration) *Typewriter {
	tw := &Typewriter{
		fullText: text,
		charRate: charRate,
	}
	if charRate <= 0 {
		tw.Skip()
	}
	return tw
}

func (tw *Typewriter) Start(now time.Time) {
	if !tw.started {
		tw.start = now
		tw.started = true
	}
}

func (tw *Typewriter) Update(now time.Time) {
	if tw.skipped || !tw.started {
		return
	}
	elapsed := now.Sub(tw.start)
	tw.revealed = int(elapsed / tw.charRate)
	max := len(tw.fullText) + len(fadeColors)
	if tw.revealed > max {
		tw.revealed = max
	}
}

func (tw *Typewriter) View() string {
	if tw.skipped || !tw.started {
		if tw.skipped {
			return tw.fullText
		}
		return ""
	}

	fadeLen := len(fadeColors)
	if tw.revealed <= 0 {
		return ""
	}

	solidEnd := tw.revealed - fadeLen
	if solidEnd < 0 {
		solidEnd = 0
	}

	result := tw.fullText[:solidEnd]

	for i := solidEnd; i < tw.revealed && i < len(tw.fullText); i++ {
		fadeIdx := i - solidEnd
		ch := tw.fullText[i]
		if ch == ' ' || ch == '\n' {
			result += string(ch)
		} else {
			style := lipgloss.NewStyle().Foreground(fadeColors[fadeIdx])
			result += style.Render(string(ch))
		}
	}

	return result
}

func (tw *Typewriter) Skip() {
	tw.skipped = true
	tw.revealed = len(tw.fullText)
}

func (tw *Typewriter) Done() bool {
	return tw.skipped || tw.revealed >= len(tw.fullText)+len(fadeColors)
}
