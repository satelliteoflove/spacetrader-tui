package screens

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/the4ofus/spacetrader-tui/internal/formula"
	"github.com/the4ofus/spacetrader-tui/internal/game"
)

const (
	chartWidth  = 70
	chartHeight = 28
)

type GalacticChartScreen struct {
	gs       *game.GameState
	cursor   int
	byDist   []int
	maxX     int
	maxY     int
}

func NewGalacticChartScreen(gs *game.GameState) *GalacticChartScreen {
	maxX, maxY := 0, 0
	for _, sys := range gs.Data.Systems {
		if sys.X > maxX {
			maxX = sys.X
		}
		if sys.Y > maxY {
			maxY = sys.Y
		}
	}

	cur := gs.Data.Systems[gs.CurrentSystemID]
	byDist := make([]int, len(gs.Data.Systems))
	for i := range byDist {
		byDist[i] = i
	}
	sort.Slice(byDist, func(i, j int) bool {
		a := gs.Data.Systems[byDist[i]]
		b := gs.Data.Systems[byDist[j]]
		da := formula.Distance(cur.X, cur.Y, a.X, a.Y)
		db := formula.Distance(cur.X, cur.Y, b.X, b.Y)
		return da < db
	})

	return &GalacticChartScreen{
		gs:   gs,
		byDist: byDist,
		maxX: maxX,
		maxY: maxY,
	}
}

func (s *GalacticChartScreen) Init() tea.Cmd { return nil }

func (s *GalacticChartScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, Keys.Up), key.Matches(msg, Keys.Down):
			delta := 1
			if key.Matches(msg, Keys.Up) {
				delta = -1
			}
			s.cursor = wrapCursor(s.cursor, delta, len(s.byDist))
		case msg.String() == "l":
			return s, func() tea.Msg { return NavigateMsg{Screen: ScreenGalacticList} }
		case key.Matches(msg, Keys.Back):
			return s, func() tea.Msg { return NavigateMsg{Screen: ScreenSystem} }
		}
	}
	return s, nil
}

func (s *GalacticChartScreen) View() string {
	var b strings.Builder

	cur := s.gs.Data.Systems[s.gs.CurrentSystemID]
	selIdx := s.byDist[s.cursor]
	sel := s.gs.Data.Systems[selIdx]

	b.WriteString(HeaderStyle.Render("  GALACTIC CHART  ") + "\n")

	selDist := formula.Distance(cur.X, cur.Y, sel.X, sel.Y)
	b.WriteString(fmt.Sprintf("  You: %s  |  Selected: %s (%.1f parsecs)\n\n",
		cur.Name, sel.Name, selDist))

	type cell struct {
		ch    rune
		style int
		label string
	}
	const (
		styleNone = iota
		styleCurrent
		styleSelected
		styleVisited
		styleInRange
		styleUnvisited
		styleWormhole
	)

	grid := make([][]cell, chartHeight)
	for y := range grid {
		grid[y] = make([]cell, chartWidth)
		for x := range grid[y] {
			grid[y][x] = cell{ch: ' ', style: styleNone}
		}
	}

	sysPos := make([][2]int, len(s.gs.Data.Systems))
	fuelRange := float64(s.gs.Player.Ship.Fuel)

	for i, sys := range s.gs.Data.Systems {
		px := sys.X * (chartWidth - 2) / (s.maxX + 1)
		py := sys.Y * (chartHeight - 1) / (s.maxY + 1)
		if px >= chartWidth {
			px = chartWidth - 1
		}
		if py >= chartHeight {
			py = chartHeight - 1
		}
		sysPos[i] = [2]int{px, py}

		dist := formula.Distance(cur.X, cur.Y, sys.X, sys.Y)

		st := styleUnvisited
		ch := rune('.')
		if s.gs.Systems[i].Visited {
			st = styleVisited
			ch = 'o'
		}
		if dist <= fuelRange {
			st = styleInRange
			ch = 'o'
		}
		if i == s.gs.CurrentSystemID {
			st = styleCurrent
			ch = '@'
		}
		if i == selIdx && i != s.gs.CurrentSystemID {
			st = styleSelected
			ch = '+'
		}

		grid[py][px] = cell{ch: ch, style: st}
	}

	labelSystem := func(sysIdx int, labelStyle int) {
		name := s.gs.Data.Systems[sysIdx].Name
		px, py := sysPos[sysIdx][0], sysPos[sysIdx][1]

		lx := px + 2
		if lx+len(name) >= chartWidth {
			lx = px - len(name) - 1
			if lx < 0 {
				lx = 0
			}
		}

		for ci, ch := range name {
			x := lx + ci
			if x >= 0 && x < chartWidth && grid[py][x].ch == ' ' {
				grid[py][x] = cell{ch: ch, style: labelStyle}
			}
		}
	}

	labelSystem(s.gs.CurrentSystemID, styleCurrent)
	if selIdx != s.gs.CurrentSystemID {
		labelSystem(selIdx, styleSelected)
	}

	for _, wh := range s.gs.Wormholes {
		isRelevant := wh.SystemA == s.gs.CurrentSystemID || wh.SystemB == s.gs.CurrentSystemID ||
			wh.SystemA == selIdx || wh.SystemB == selIdx
		if isRelevant {
			labelSystem(wh.SystemA, styleWormhole)
			labelSystem(wh.SystemB, styleWormhole)
		}
	}

	for y := range grid {
		var line strings.Builder
		line.WriteString("  ")
		for x := range grid[y] {
			c := grid[y][x]
			s := string(c.ch)
			switch c.style {
			case styleCurrent:
				line.WriteString(SelectedStyle.Render(s))
			case styleSelected:
				line.WriteString("\033[36m" + s + "\033[0m")
			case styleVisited:
				line.WriteString(DimStyle.Render(s))
			case styleInRange:
				line.WriteString(SuccessStyle.Render(s))
			case styleWormhole:
				line.WriteString("\033[35m" + s + "\033[0m")
			default:
				line.WriteString(s)
			}
		}
		b.WriteString(line.String() + "\n")
	}

	b.WriteString("\n")
	b.WriteString(DimStyle.Render("  @ you  + selected  green = in range  magenta = wormhole") + "\n")

	selState := s.gs.Systems[selIdx]
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("  %s", "\033[36m"+sel.Name+"\033[0m"))
	b.WriteString(fmt.Sprintf("  |  Tech: %s  |  Gov: %s", shortTech(sel.TechLevel), sel.PoliticalSystem))
	if sel.Resource.String() != "No Special Resources" {
		b.WriteString(fmt.Sprintf("  |  %s", shortResource(sel.Resource)))
	}
	b.WriteString("\n")

	if selState.Visited {
		if selState.Event != "" {
			b.WriteString(DangerStyle.Render(fmt.Sprintf("  Event: %s", selState.Event)) + "\n")
		}
	} else {
		b.WriteString(DimStyle.Render("  Not yet visited") + "\n")
	}

	for _, wh := range s.gs.Wormholes {
		if wh.SystemA == selIdx {
			b.WriteString(SuccessStyle.Render(fmt.Sprintf("  Wormhole to %s", s.gs.Data.Systems[wh.SystemB].Name)) + "\n")
		} else if wh.SystemB == selIdx {
			b.WriteString(SuccessStyle.Render(fmt.Sprintf("  Wormhole to %s", s.gs.Data.Systems[wh.SystemA].Name)) + "\n")
		}
	}

	b.WriteString("\n" + DimStyle.Render("  j/k cycle systems, l = list view, esc back"))
	return b.String()
}

type GalacticListScreen struct {
	gs     *game.GameState
	cursor int
}

func NewGalacticListScreen(gs *game.GameState) *GalacticListScreen {
	return &GalacticListScreen{gs: gs}
}

func (s *GalacticListScreen) Init() tea.Cmd { return nil }

func (s *GalacticListScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, Keys.Up):
			s.cursor = wrapCursor(s.cursor, -1, len(s.gs.Data.Systems))
		case key.Matches(msg, Keys.Down):
			s.cursor = wrapCursor(s.cursor, 1, len(s.gs.Data.Systems))
		case msg.String() == "m":
			return s, func() tea.Msg { return NavigateMsg{Screen: ScreenGalacticChart} }
		case key.Matches(msg, Keys.Back):
			return s, func() tea.Msg { return NavigateMsg{Screen: ScreenSystem} }
		}
	}
	return s, nil
}

func (s *GalacticListScreen) View() string {
	var b strings.Builder

	b.WriteString(HeaderStyle.Render("  ALL SYSTEMS  ") + "\n\n")

	cur := s.gs.Data.Systems[s.gs.CurrentSystemID]

	pageSize := 15
	start := s.cursor - pageSize/2
	if start < 0 {
		start = 0
	}
	end := start + pageSize
	if end > len(s.gs.Data.Systems) {
		end = len(s.gs.Data.Systems)
		start = end - pageSize
		if start < 0 {
			start = 0
		}
	}

	b.WriteString(fmt.Sprintf("  %-16s %5s  %-10s %-16s %-8s\n",
		"SYSTEM", "DIST", "TECH", "GOVERNMENT", "RESOURCE"))
	b.WriteString("  " + strings.Repeat("-", 60) + "\n")

	for i := start; i < end; i++ {
		sys := s.gs.Data.Systems[i]
		dist := math.Sqrt(float64((cur.X-sys.X)*(cur.X-sys.X) + (cur.Y-sys.Y)*(cur.Y-sys.Y)))

		marker := " "
		if s.gs.Systems[i].Visited {
			marker = "*"
		}
		if i == s.gs.CurrentSystemID {
			marker = "@"
		}

		resStr := shortResource(sys.Resource)

		line := fmt.Sprintf("%-16s %5.1f  %-10s %-16s %-8s %s",
			sys.Name, dist, shortTech(sys.TechLevel), sys.PoliticalSystem, resStr, marker)

		if i == s.cursor {
			b.WriteString(SelectedStyle.Render("> ") + line + "\n")
		} else {
			b.WriteString("  " + line + "\n")
		}
	}

	b.WriteString(fmt.Sprintf("\n  System %d of %d\n", s.cursor+1, len(s.gs.Data.Systems)))

	if s.cursor < len(s.gs.Data.Systems) {
		sys := s.gs.Data.Systems[s.cursor]
		sysState := s.gs.Systems[s.cursor]
		b.WriteString("\n" + DimStyle.Render(fmt.Sprintf("  --- %s ---", sys.Name)) + "\n")
		b.WriteString(DimStyle.Render(fmt.Sprintf("  Coordinates: (%d, %d)", sys.X, sys.Y)) + "\n")
		if sysState.Visited {
			if sysState.Event != "" {
				b.WriteString(DangerStyle.Render(fmt.Sprintf("  Event: %s", sysState.Event)) + "\n")
			}
		} else {
			b.WriteString(DimStyle.Render("  Not yet visited") + "\n")
		}
		for _, wh := range s.gs.Wormholes {
			if wh.SystemA == s.cursor {
				b.WriteString(SuccessStyle.Render(fmt.Sprintf("  Wormhole to %s", s.gs.Data.Systems[wh.SystemB].Name)) + "\n")
			} else if wh.SystemB == s.cursor {
				b.WriteString(SuccessStyle.Render(fmt.Sprintf("  Wormhole to %s", s.gs.Data.Systems[wh.SystemA].Name)) + "\n")
			}
		}
	}

	b.WriteString("\n" + DimStyle.Render("  j/k scroll, m = map view, esc back"))
	return b.String()
}
