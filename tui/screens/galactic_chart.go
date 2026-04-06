package screens

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/the4ofus/spacetrader-tui/internal/formula"
	"github.com/the4ofus/spacetrader-tui/internal/game"
	"github.com/the4ofus/spacetrader-tui/internal/travel"
)

const (
	chartWidth  = 70
	chartHeight = 28
)

type GalacticChartScreen struct {
	gs          *game.GameState
	curX        int
	curY        int
	sysAt       map[[2]int]int
	sysPos      [][2]int
	maxX        int
	maxY        int
	searchMode  bool
	searchInput textinput.Model
	message     string
	confirming  bool
}

func NewGalacticChartScreen(gs *game.GameState) *GalacticChartScreen {
	return NewGalacticChartScreenWithSelection(gs, -1)
}

func NewGalacticChartScreenWithSelection(gs *game.GameState, selectedSys int) *GalacticChartScreen {
	maxX, maxY := 0, 0
	for _, sys := range gs.Data.Systems {
		if sys.X > maxX {
			maxX = sys.X
		}
		if sys.Y > maxY {
			maxY = sys.Y
		}
	}

	sysAt := make(map[[2]int]int)
	sysPos := make([][2]int, len(gs.Data.Systems))
	for i, sys := range gs.Data.Systems {
		px := sys.X * (chartWidth - 2) / (maxX + 1)
		py := sys.Y * (chartHeight - 1) / (maxY + 1)
		if px >= chartWidth {
			px = chartWidth - 1
		}
		if py >= chartHeight {
			py = chartHeight - 1
		}
		sysPos[i] = [2]int{px, py}
		sysAt[[2]int{px, py}] = i
	}

	startSys := selectedSys
	if startSys < 0 {
		startSys = gs.CurrentSystemID
	}

	return &GalacticChartScreen{
		gs:          gs,
		curX:        sysPos[startSys][0],
		curY:        sysPos[startSys][1],
		sysAt:       sysAt,
		sysPos:      sysPos,
		maxX:        maxX,
		maxY:        maxY,
		searchInput: newFilterInput(),
	}
}

func (s *GalacticChartScreen) selectedSystem() (int, bool) {
	idx, ok := s.sysAt[[2]int{s.curX, s.curY}]
	return idx, ok
}

func (s *GalacticChartScreen) Init() tea.Cmd { return nil }

func (s *GalacticChartScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if s.searchMode {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			if key.Matches(msg, Keys.Enter) {
				s.searchMode = false
				query := strings.ToLower(s.searchInput.Value())
				if query != "" {
					for i, sys := range s.gs.Data.Systems {
						if strings.Contains(strings.ToLower(sys.Name), query) {
							s.curX = s.sysPos[i][0]
							s.curY = s.sysPos[i][1]
							break
						}
					}
				}
				return s, nil
			}
			if key.Matches(msg, Keys.Back) {
				s.searchMode = false
				return s, nil
			}
		}
		var cmd tea.Cmd
		s.searchInput, cmd = s.searchInput.Update(msg)
		return s, cmd
	}

	if s.confirming {
		if msg, ok := msg.(tea.KeyMsg); ok {
			switch msg.String() {
			case "y":
				s.confirming = false
				if selIdx, ok := s.selectedSystem(); ok {
					result := travel.ExecuteTravel(s.gs, selIdx)
					if !result.Success {
						s.message = result.Message
						return s, nil
					}
					s.message = result.Message
					return s, func() tea.Msg { return TravelMsg{DestIdx: selIdx} }
				}
			default:
				s.confirming = false
				s.message = ""
			}
		}
		return s, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case msg.String() == "up" || msg.String() == "k":
			if s.curY > 0 {
				s.curY--
			}
		case msg.String() == "down" || msg.String() == "j":
			if s.curY < chartHeight-1 {
				s.curY++
			}
		case msg.String() == "left" || msg.String() == "h":
			if s.curX > 0 {
				s.curX--
			}
		case msg.String() == "right" || msg.String() == "l":
			if s.curX < chartWidth-1 {
				s.curX++
			}
		case msg.String() == "b":
			if selIdx, ok := s.selectedSystem(); ok {
				s.gs.ToggleBookmark(selIdx, autoBookmarkNote(s.gs, selIdx))
			}
		case msg.String() == "/":
			s.searchMode = true
			s.searchInput.SetValue("")
			s.searchInput.Focus()
			return s, textinput.Blink
		case key.Matches(msg, Keys.Enter):
			if selIdx, ok := s.selectedSystem(); ok {
				cur := s.gs.Data.Systems[s.gs.CurrentSystemID]
				dest := s.gs.Data.Systems[selIdx]
				dist := formula.Distance(cur.X, cur.Y, dest.X, dest.Y)
				if dist > float64(s.gs.Player.Ship.Fuel) {
					s.message = DangerStyle.Render(fmt.Sprintf("Out of range (%.1f parsecs, fuel: %d)", dist, s.gs.Player.Ship.Fuel))
				} else {
					s.message = SelectedStyle.Render(fmt.Sprintf("Travel to %s? (y/n)", dest.Name))
					s.confirming = true
				}
			}
		case msg.String() == "L":
			selIdx := -1
			if idx, ok := s.selectedSystem(); ok {
				selIdx = idx
			}
			return s, func() tea.Msg { return NavigateMsg{Screen: ScreenGalacticList, SelectedSystem: selIdx} }
		case key.Matches(msg, Keys.Back):
			return s, func() tea.Msg { return NavigateMsg{Screen: ScreenSystem} }
		}
	}
	return s, nil
}

func (s *GalacticChartScreen) View() string {
	var b strings.Builder

	cur := s.gs.Data.Systems[s.gs.CurrentSystemID]
	selIdx, hasSel := s.selectedSystem()

	b.WriteString(HeaderStyle.Render("  GALACTIC CHART  ") + "\n")

	if hasSel {
		sel := s.gs.Data.Systems[selIdx]
		selDist := formula.Distance(cur.X, cur.Y, sel.X, sel.Y)
		b.WriteString(fmt.Sprintf("  You: %s  |  Cursor: %s (%.1f parsecs)\n\n",
			cur.Name, sel.Name, selDist))
	} else {
		b.WriteString(fmt.Sprintf("  You: %s  |  Cursor: (%d, %d)\n\n",
			cur.Name, s.curX, s.curY))
	}

	type cell struct {
		ch    rune
		style int
	}
	const (
		styleNone = iota
		styleCurrent
		styleSelected
		styleCursor
		styleVisited
		styleInRange
		styleUnvisited
		styleWormhole
		styleBookmarked
	)

	grid := make([][]cell, chartHeight)
	for y := range grid {
		grid[y] = make([]cell, chartWidth)
		for x := range grid[y] {
			grid[y][x] = cell{ch: ' ', style: styleNone}
		}
	}

	fuelRange := float64(s.gs.Player.Ship.Fuel)

	for i, sys := range s.gs.Data.Systems {
		px, py := s.sysPos[i][0], s.sysPos[i][1]

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
		if s.gs.IsBookmarked(i) {
			st = styleBookmarked
			ch = '!'
		}
		if i == s.gs.CurrentSystemID {
			st = styleCurrent
			ch = '@'
		}
		if hasSel && i == selIdx && i != s.gs.CurrentSystemID {
			st = styleSelected
			ch = '+'
		}

		grid[py][px] = cell{ch: ch, style: st}
	}

	if !hasSel && s.curY >= 0 && s.curY < chartHeight && s.curX >= 0 && s.curX < chartWidth {
		grid[s.curY][s.curX] = cell{ch: 'x', style: styleCursor}
	}

	labelSystem := func(sysIdx int, labelStyle int) {
		name := s.gs.Data.Systems[sysIdx].Name
		px, py := s.sysPos[sysIdx][0], s.sysPos[sysIdx][1]

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
	if hasSel && selIdx != s.gs.CurrentSystemID {
		labelSystem(selIdx, styleSelected)
	}

	for _, wh := range s.gs.Wormholes {
		isRelevant := wh.SystemA == s.gs.CurrentSystemID || wh.SystemB == s.gs.CurrentSystemID
		if hasSel {
			isRelevant = isRelevant || wh.SystemA == selIdx || wh.SystemB == selIdx
		}
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
			case styleCursor:
				line.WriteString(DimStyle.Render(s))
			case styleVisited:
				line.WriteString(DimStyle.Render(s))
			case styleInRange:
				line.WriteString(SuccessStyle.Render(s))
			case styleWormhole:
				line.WriteString(MagentaStyle.Render(s))
			case styleBookmarked:
				line.WriteString(SelectedStyle.Render(s))
			default:
				line.WriteString(s)
			}
		}
		b.WriteString(line.String() + "\n")
	}

	b.WriteString("\n")
	b.WriteString(DimStyle.Render("  @ you  + selected  ! bookmarked  green = in range") + "\n")

	if hasSel {
		sel := s.gs.Data.Systems[selIdx]
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

		if bm, ok := s.gs.GetBookmark(selIdx); ok && bm.Note != "" {
			b.WriteString(SelectedStyle.Render(fmt.Sprintf("  Bookmarked: %s", bm.Note)) + "\n")
		}
	}

	if s.message != "" {
		b.WriteString("\n  " + s.message + "\n")
	}

	if s.searchMode {
		b.WriteString("  / " + s.searchInput.View() + "\n")
	}

	b.WriteString("\n" + DimStyle.Render("  arrows/hjkl move, / search, enter travel, b bookmark, L list, esc back"))
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
	return NewGalacticListScreenWithSelection(gs, -1)
}

func NewGalacticListScreenWithSelection(gs *game.GameState, selectedSys int) *GalacticListScreen {
	entries := buildAllSystemEntries(gs)
	filtered := applyFilterAndSort(entries, "", colName, sortAsc)
	cursor := 0
	if selectedSys >= 0 {
		for i, e := range filtered {
			if e.sysIdx == selectedSys {
				cursor = i
				break
			}
		}
	}
	return &GalacticListScreen{
		gs:          gs,
		cursor:      cursor,
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
		case msg.String() == "b":
			if len(s.filtered) > 0 {
				entry := s.filtered[s.cursor]
				s.gs.ToggleBookmark(entry.sysIdx, autoBookmarkNote(s.gs, entry.sysIdx))
				s.allEntries = buildAllSystemEntries(s.gs)
				s.refilter()
			}
		case msg.String() == "m":
			sysIdx := -1
			if len(s.filtered) > 0 {
				sysIdx = s.filtered[s.cursor].sysIdx
			}
			return s, func() tea.Msg { return NavigateMsg{Screen: ScreenGalacticChart, SelectedSystem: sysIdx} }
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
	resH := sortedHeader("SPECIALTY", colResource, s.sortCol, s.sortDir)

	header := fmt.Sprintf("  %-16s %5s  %-10s %-16s %-12s",
		sysH, distH, techH, govH, resH)
	b.WriteString(DimStyle.Render(header) + "\n")
	b.WriteString("  " + strings.Repeat("-", 64) + "\n")

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
			if e.bookmarked {
				marker = SelectedStyle.Render("!")
			} else if e.isCurrent {
				marker = "@"
			} else if e.visited {
				marker = "*"
			}

			coloredRes := colorResource(e.resource, fmt.Sprintf("%-12s", e.resStr))
			line := fmt.Sprintf("%-16s %5.1f  %-10s %-16s",
				e.name, e.dist, e.techStr, e.govStr)
			line += coloredRes + " " + marker

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
		if e.bookmarked && e.bookmarkNote != "" {
			b.WriteString(SelectedStyle.Render(fmt.Sprintf("  Bookmarked: %s", e.bookmarkNote)) + "\n")
		}
	}

	b.WriteString("\n" + DimStyle.Render("  j/k scroll, b bookmark, 1-5 sort, / filter, m map, esc back"))
	return b.String()
}
