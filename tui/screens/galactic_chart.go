package screens

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
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
			if x >= 0 && x < chartWidth {
				existing := grid[py][x]
				if existing.ch == ' ' || existing.style == styleUnvisited ||
					existing.style == styleVisited || existing.style == styleInRange {
					grid[py][x] = cell{ch: ch, style: labelStyle}
				}
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
				line.WriteString(CyanStyle.Render(s))
			case styleVisited:
				line.WriteString(DimStyle.Render(s))
			case styleInRange:
				line.WriteString(SuccessStyle.Render(s))
			case styleWormhole:
				line.WriteString(MagentaStyle.Render(s))
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
	b.WriteString(fmt.Sprintf("  %s", CyanStyle.Render(sel.Name)))
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
	gs          *game.GameState
	cursor      int
	allEntries  []systemEntry
	filtered    []systemEntry
	sortCol     sortColumn
	sortDir     sortDir
	filterMode  bool
	filterInput textinput.Model
	filterText  string
}

func NewGalacticListScreen(gs *game.GameState) *GalacticListScreen {
	entries := buildAllSystemEntries(gs)
	filtered := applyFilterAndSort(entries, "", colName, sortAsc)
	return &GalacticListScreen{
		gs:          gs,
		allEntries:  entries,
		filtered:    filtered,
		sortCol:     colName,
		sortDir:     sortAsc,
		filterInput: newFilterInput(),
	}
}

func (s *GalacticListScreen) Init() tea.Cmd { return nil }

func (s *GalacticListScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if s.filterMode {
		return s.updateFilter(msg)
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, Keys.Up):
			if len(s.filtered) > 0 {
				s.cursor = wrapCursor(s.cursor, -1, len(s.filtered))
			}
		case key.Matches(msg, Keys.Down):
			if len(s.filtered) > 0 {
				s.cursor = wrapCursor(s.cursor, 1, len(s.filtered))
			}
		case msg.String() == "1":
			s.toggleSort(colName)
		case msg.String() == "2":
			s.toggleSort(colDist)
		case msg.String() == "3":
			s.toggleSort(colTech)
		case msg.String() == "4":
			s.toggleSort(colGov)
		case msg.String() == "5":
			s.toggleSort(colResource)
		case msg.String() == "/":
			s.filterMode = true
			s.filterInput.SetValue(s.filterText)
			s.filterInput.Focus()
			return s, textinput.Blink
		case msg.String() == "m":
			return s, func() tea.Msg { return NavigateMsg{Screen: ScreenGalacticChart} }
		case key.Matches(msg, Keys.Back):
			return s, func() tea.Msg { return NavigateMsg{Screen: ScreenSystem} }
		}
	}
	return s, nil
}

func (s *GalacticListScreen) updateFilter(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if key.Matches(msg, Keys.Enter) {
			s.filterMode = false
			s.filterText = s.filterInput.Value()
			s.refilter()
			return s, nil
		}
		if key.Matches(msg, Keys.Back) {
			s.filterMode = false
			s.filterText = ""
			s.filterInput.SetValue("")
			s.refilter()
			return s, nil
		}
	}
	var cmd tea.Cmd
	s.filterInput, cmd = s.filterInput.Update(msg)
	s.filterText = s.filterInput.Value()
	s.refilter()
	return s, cmd
}

func (s *GalacticListScreen) toggleSort(col sortColumn) {
	if s.sortCol == col {
		if s.sortDir == sortAsc {
			s.sortDir = sortDesc
		} else {
			s.sortDir = sortAsc
		}
	} else {
		s.sortCol = col
		s.sortDir = sortAsc
	}
	s.refilter()
}

func (s *GalacticListScreen) refilter() {
	s.filtered = applyFilterAndSort(s.allEntries, s.filterText, s.sortCol, s.sortDir)
	if s.cursor >= len(s.filtered) {
		if len(s.filtered) > 0 {
			s.cursor = len(s.filtered) - 1
		} else {
			s.cursor = 0
		}
	}
}

func (s *GalacticListScreen) View() string {
	var b strings.Builder

	b.WriteString(HeaderStyle.Render("  ALL SYSTEMS  ") + "\n")

	if s.filterMode {
		b.WriteString("  / " + s.filterInput.View() + "\n")
	} else if s.filterText != "" {
		b.WriteString(DimStyle.Render(fmt.Sprintf("  filter: %s  (/ edit, esc clear)", s.filterText)) + "\n")
	}

	b.WriteString("\n")

	sysH := sortedHeader("SYSTEM", colName, s.sortCol, s.sortDir)
	distH := sortedHeader("DIST", colDist, s.sortCol, s.sortDir)
	techH := sortedHeader("TECH", colTech, s.sortCol, s.sortDir)
	govH := sortedHeader("GOV", colGov, s.sortCol, s.sortDir)
	resH := sortedHeader("RESOURCE", colResource, s.sortCol, s.sortDir)

	header := fmt.Sprintf("  %-16s %5s  %-10s %-16s %-8s",
		sysH, distH, techH, govH, resH)
	b.WriteString(DimStyle.Render(header) + "\n")
	b.WriteString("  " + strings.Repeat("-", 60) + "\n")

	if len(s.filtered) == 0 {
		b.WriteString("  No matching systems.\n")
	} else {
		pageSize := 15
		start := s.cursor - pageSize/2
		if start < 0 {
			start = 0
		}
		end := start + pageSize
		if end > len(s.filtered) {
			end = len(s.filtered)
			start = end - pageSize
			if start < 0 {
				start = 0
			}
		}

		for i := start; i < end; i++ {
			e := s.filtered[i]

			marker := " "
			if e.visited {
				marker = "*"
			}
			if e.isCurrent {
				marker = "@"
			}

			line := fmt.Sprintf("%-16s %5.1f  %-10s %-16s %-8s %s",
				e.name, e.dist, e.techStr, e.govStr, e.resStr, marker)

			if i == s.cursor {
				b.WriteString(SelectedStyle.Render("> ") + line + "\n")
			} else {
				b.WriteString("  " + line + "\n")
			}
		}

		countStr := fmt.Sprintf("  System %d of %d", s.cursor+1, len(s.filtered))
		if s.filterText != "" {
			countStr += fmt.Sprintf(" (%d total)", len(s.allEntries))
		}
		b.WriteString("\n" + countStr + "\n")
	}

	if s.cursor < len(s.filtered) {
		e := s.filtered[s.cursor]
		sys := s.gs.Data.Systems[e.sysIdx]
		sysState := s.gs.Systems[e.sysIdx]

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
			if wh.SystemA == e.sysIdx {
				b.WriteString(SuccessStyle.Render(fmt.Sprintf("  Wormhole to %s", s.gs.Data.Systems[wh.SystemB].Name)) + "\n")
			} else if wh.SystemB == e.sysIdx {
				b.WriteString(SuccessStyle.Render(fmt.Sprintf("  Wormhole to %s", s.gs.Data.Systems[wh.SystemA].Name)) + "\n")
			}
		}
	}

	b.WriteString("\n" + DimStyle.Render("  j/k scroll, 1-5 sort, / filter, m map, esc back"))
	return b.String()
}
