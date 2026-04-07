package screens

import (
	"math/rand"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	warpWidth  = 60
	warpHeight = 16
)

type WarpScreen struct {
	stars    [warpHeight][warpWidth]rune
	frame    int
	destName string
}

func NewWarpScreen(destName string) *WarpScreen {
	w := &WarpScreen{destName: destName}
	for y := 0; y < warpHeight; y++ {
		spawnStarRow(&w.stars[y])
	}
	return w
}

func spawnStarRow(row *[warpWidth]rune) {
	for x := 0; x < warpWidth; x++ {
		row[x] = ' '
	}
	count := 3 + rand.Intn(4)
	for i := 0; i < count; i++ {
		x := rand.Intn(warpWidth)
		if rand.Intn(3) == 0 {
			row[x] = '*'
		} else {
			row[x] = '.'
		}
	}
}

func (s *WarpScreen) Init() tea.Cmd { return nil }

func (s *WarpScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case TickMsg:
		s.frame++
		if AnimWarpMaxFrames <= 0 || s.frame >= AnimWarpMaxFrames {
			return s, func() tea.Msg { return WarpDoneMsg{} }
		}
		for y := warpHeight - 1; y > 0; y-- {
			s.stars[y] = s.stars[y-1]
		}
		spawnStarRow(&s.stars[0])
	}
	return s, nil
}

func (s *WarpScreen) View() string {
	var b strings.Builder

	brightStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("15")).Bold(true)
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("14")).Bold(true)

	b.WriteString("\n")
	b.WriteString(labelStyle.Render("  Warping to " + s.destName + "..."))
	b.WriteString("\n\n")

	for y := 0; y < warpHeight; y++ {
		for x := 0; x < warpWidth; x++ {
			ch := s.stars[y][x]
			switch ch {
			case '*':
				b.WriteString(brightStyle.Render("*"))
			case '.':
				b.WriteString(dimStyle.Render("."))
			default:
				b.WriteByte(' ')
			}
		}
		b.WriteByte('\n')
	}

	return b.String()
}
