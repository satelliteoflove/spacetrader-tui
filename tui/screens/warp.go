package screens

import (
	"fmt"
	"math/rand"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/the4ofus/spacetrader-tui/internal/encounter"
	"github.com/the4ofus/spacetrader-tui/internal/game"
)

const (
	warpWidth  = 60
	warpHeight = 16
)

type WarpScreen struct {
	gs        *game.GameState
	stars     [warpHeight][warpWidth]rune
	frame     int
	click     int
	maxClicks int
	destIdx   int
	destName  string
	done      bool
}

func NewWarpScreen(gs *game.GameState, destIdx int, destName string, distance float64) *WarpScreen {
	shipDef := gs.Data.Ships[gs.Player.Ship.TypeID]
	w := &WarpScreen{
		gs:        gs,
		destIdx:   destIdx,
		destName:  destName,
		maxClicks: encounter.ClicksForDistance(distance, shipDef.Range),
	}
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
		if s.done {
			return s, nil
		}
		s.frame++

		for y := warpHeight - 1; y > 0; y-- {
			s.stars[y] = s.stars[y-1]
		}
		spawnStarRow(&s.stars[0])

		shouldAdvance := AnimWarpMaxFrames <= 0 || s.frame%max(AnimWarpMaxFrames/s.maxClicks, 1) == 0
		if shouldAdvance {
			if cmd := s.advanceClick(); cmd != nil {
				return s, cmd
			}
			if AnimWarpMaxFrames <= 0 {
				s.done = true
				return s, func() tea.Msg { return WarpDoneMsg{} }
			}
		}

	case WarpResumeMsg:
		if s.gs.EndStatus == game.StatusDead {
			s.done = true
			return s, func() tea.Msg { return WarpDoneMsg{} }
		}
	}
	return s, nil
}

func (s *WarpScreen) advanceClick() tea.Cmd {
	s.click++
	if s.click > s.maxClicks {
		s.done = true
		return func() tea.Msg { return WarpDoneMsg{} }
	}
	if enc := encounter.GenerateForClick(s.gs, s.destIdx); enc != nil {
		return func() tea.Msg { return WarpEncounterMsg{Encounter: enc} }
	}
	return nil
}

func (s *WarpScreen) View() string {
	var b strings.Builder

	brightStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("15")).Bold(true)
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("14")).Bold(true)

	b.WriteString("\n")
	b.WriteString(labelStyle.Render("  Warping to " + s.destName + "..."))

	progress := s.click * 100 / s.maxClicks
	if progress > 100 {
		progress = 100
	}
	b.WriteString(dimStyle.Render(fmt.Sprintf("  [%d%%]", progress)))
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
