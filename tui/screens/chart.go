package screens

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/the4ofus/spacetrader-tui/internal/game"
	"github.com/the4ofus/spacetrader-tui/internal/gamedata"
	"github.com/the4ofus/spacetrader-tui/internal/shipyard"
	"github.com/the4ofus/spacetrader-tui/internal/travel"
)

type ChartScreen struct {
	gs          *game.GameState
	cursor      int
	allEntries  []systemEntry
	filtered    []systemEntry
	sortCol     sortColumn
	sortDir     sortDir
	filterMode  bool
	filterInput textinput.Model
	filterText  string
	message     string
	confirming  bool
}

func NewChartScreen(gs *game.GameState) *ChartScreen {
	systems := travel.ReachableSystems(gs)
	indices := make([]int, len(systems))
	for i, rs := range systems {
		indices[i] = rs.Index
	}
	allEntries := buildSystemEntries(gs, indices)
	filtered := applyFilterAndSort(allEntries, "", colDist, sortAsc)
	return &ChartScreen{
		gs:          gs,
		allEntries:  allEntries,
		filtered:    filtered,
		sortCol:     colDist,
		sortDir:     sortAsc,
		filterInput: newFilterInput(),
	}
}

func (s *ChartScreen) Init() tea.Cmd { return nil }

func (s *ChartScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if s.filterMode {
		return s.updateChartFilter(msg)
	}

	if s.confirming {
		if msg, ok := msg.(tea.KeyMsg); ok {
			switch msg.String() {
			case "y":
				s.confirming = false
				entry := s.filtered[s.cursor]
				result := travel.ExecuteTravel(s.gs, entry.sysIdx)
				if !result.Success {
					s.message = result.Message
					return s, nil
				}
				s.message = result.Message
				return s, func() tea.Msg { return TravelMsg{DestIdx: entry.sysIdx} }
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
		case key.Matches(msg, Keys.Up):
			if len(s.filtered) > 0 {
				s.cursor = wrapCursor(s.cursor, -1, len(s.filtered))
			}
		case key.Matches(msg, Keys.Down):
			if len(s.filtered) > 0 {
				s.cursor = wrapCursor(s.cursor, 1, len(s.filtered))
			}
		case key.Matches(msg, Keys.Enter):
			if len(s.filtered) == 0 {
				return s, nil
			}
			entry := s.filtered[s.cursor]
			s.message = SelectedStyle.Render(fmt.Sprintf("Travel to %s? (y/n)", entry.name))
			s.confirming = true
		case msg.String() == "1":
			s.toggleChartSort(colName)
		case msg.String() == "2":
			s.toggleChartSort(colDist)
		case msg.String() == "3":
			s.toggleChartSort(colTech)
		case msg.String() == "4":
			s.toggleChartSort(colGov)
		case msg.String() == "5":
			s.toggleChartSort(colResource)
		case msg.String() == "/":
			s.filterMode = true
			s.filterInput.SetValue(s.filterText)
			s.filterInput.Focus()
			return s, textinput.Blink
		case msg.String() == "b":
			if len(s.filtered) > 0 {
				entry := s.filtered[s.cursor]
				added := s.gs.ToggleBookmark(entry.sysIdx, autoBookmarkNote(s.gs, entry.sysIdx))
				if added {
					s.message = SuccessStyle.Render(fmt.Sprintf("Bookmarked %s", entry.name))
				} else {
					s.message = DimStyle.Render(fmt.Sprintf("Removed bookmark for %s", entry.name))
				}
				s.refreshChartSystems()
			}
		case msg.String() == "r":
			result := shipyard.Refuel(s.gs)
			s.message = result.Message
			s.refreshChartSystems()
			return s, nil
		case msg.String() == "w":
			ok, msg := game.TravelWormhole(s.gs)
			if ok {
				s.message = msg
				return s, func() tea.Msg { return TravelMsg{DestIdx: s.gs.CurrentSystemID} }
			}
			s.message = msg
			return s, nil
		case key.Matches(msg, Keys.Back):
			return s, func() tea.Msg { return NavigateMsg{Screen: ScreenSystem} }
		}
	}
	return s, nil
}

func (s *ChartScreen) updateChartFilter(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if key.Matches(msg, Keys.Enter) {
			s.filterMode = false
			s.filterText = s.filterInput.Value()
			s.refilterChart()
			return s, nil
		}
		if key.Matches(msg, Keys.Back) {
			s.filterMode = false
			s.filterText = ""
			s.filterInput.SetValue("")
			s.refilterChart()
			return s, nil
		}
	}
	var cmd tea.Cmd
	s.filterInput, cmd = s.filterInput.Update(msg)
	s.filterText = s.filterInput.Value()
	s.refilterChart()
	return s, cmd
}

func (s *ChartScreen) toggleChartSort(col sortColumn) {
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
	s.refilterChart()
}

func (s *ChartScreen) refilterChart() {
	s.filtered = applyFilterAndSort(s.allEntries, s.filterText, s.sortCol, s.sortDir)
	if s.cursor >= len(s.filtered) {
		if len(s.filtered) > 0 {
			s.cursor = len(s.filtered) - 1
		} else {
			s.cursor = 0
		}
	}
}

func (s *ChartScreen) refreshChartSystems() {
	systems := travel.ReachableSystems(s.gs)
	indices := make([]int, len(systems))
	for i, rs := range systems {
		indices[i] = rs.Index
	}
	s.allEntries = buildSystemEntries(s.gs, indices)
	s.refilterChart()
}

func (s *ChartScreen) View() string {
	var b strings.Builder

	shipDef := s.gs.PlayerShipDef()
	cur := s.gs.CurrentSystem()

	b.WriteString(HeaderStyle.Render("  SHORT-RANGE CHART  ") + "\n")
	b.WriteString(fmt.Sprintf("  Current: %s  |  Fuel: %d/%d parsecs\n",
		cur.Name, s.gs.Player.Ship.Fuel, shipDef.Range))

	fuelCost := shipyard.RefuelCost(s.gs)
	if fuelCost > 0 {
		b.WriteString(DimStyle.Render(fmt.Sprintf("  Refuel cost: %d credits (press r)", fuelCost)) + "\n")
	}

	if dest, ok := game.WormholeDestination(s.gs, s.gs.CurrentSystemID); ok {
		destName := s.gs.Data.Systems[dest].Name
		b.WriteString(SuccessStyle.Render(fmt.Sprintf("  Wormhole to %s available! (press w)", destName)) + "\n")
	}

	if s.filterMode {
		b.WriteString("  / " + s.filterInput.View() + "\n")
	} else if s.filterText != "" {
		b.WriteString(DimStyle.Render(fmt.Sprintf("  filter: %s  (/ edit, esc clear)", s.filterText)) + "\n")
	}

	b.WriteString("\n")

	if len(s.allEntries) == 0 {
		b.WriteString("  No systems in range.\n")
		if s.gs.Player.Ship.Fuel < shipDef.Range {
			b.WriteString("  " + DimStyle.Render("Press r to refuel.") + "\n")
		}
	} else {
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
			for i, e := range s.filtered {
				marker := " "
				if e.bookmarked {
					marker = SelectedStyle.Render("!")
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
		}

		b.WriteString("\n" + DimStyle.Render("  * = visited  ! = bookmarked"))
	}

	if s.message != "" {
		b.WriteString("\n  " + s.message)
	}

	if len(s.filtered) > 0 && s.cursor < len(s.filtered) {
		e := s.filtered[s.cursor]
		destDef := s.gs.Data.Systems[e.sysIdx]
		destState := s.gs.Systems[e.sysIdx]

		b.WriteString("\n")
		b.WriteString(DimStyle.Render(fmt.Sprintf("  --- %s ---", destDef.Name)) + "\n")

		availableGoods := 0
		for g := 0; g < game.NumGoods; g++ {
			if destState.Prices[g] > 0 {
				availableGoods++
			}
		}
		if s.gs.Systems[e.sysIdx].Visited {
			b.WriteString(DimStyle.Render(fmt.Sprintf("  %d goods available  |  Event: %s",
				availableGoods, eventOrNone(destState.Event))) + "\n")
		} else {
			b.WriteString(DimStyle.Render("  Not yet visited") + "\n")
		}

		if e.resource != gamedata.ResourceNone {
			cheap, expensive := resourceTradeHints(s.gs.Data.Goods, destDef.Resource)
			if cheap != "" {
				b.WriteString(SuccessStyle.Render(fmt.Sprintf("  Buy cheap: %s", cheap)) + "\n")
			}
			if expensive != "" {
				b.WriteString(DangerStyle.Render(fmt.Sprintf("  Sells high: %s", expensive)) + "\n")
			}
		}

		if e.bookmarked && e.bookmarkNote != "" {
			b.WriteString(SelectedStyle.Render(fmt.Sprintf("  Bookmarked: %s", e.bookmarkNote)) + "\n")
		}
	}

	b.WriteString("\n" + DimStyle.Render("  enter travel, r refuel, w wormhole, b bookmark, 1-5 sort, / filter, esc back"))
	return b.String()
}

func shortTech(t gamedata.TechLevel) string {
	switch t {
	case gamedata.TechPreAgricultural:
		return "Pre-ag"
	case gamedata.TechAgricultural:
		return "Agri"
	case gamedata.TechMedieval:
		return "Medieval"
	case gamedata.TechRenaissance:
		return "Renais"
	case gamedata.TechEarlyIndustrial:
		return "Early Ind"
	case gamedata.TechIndustrial:
		return "Industrial"
	case gamedata.TechPostIndustrial:
		return "Post-ind"
	case gamedata.TechHiTech:
		return "Hi-tech"
	}
	return t.String()
}

func shortResource(r gamedata.Resource) string {
	switch r {
	case gamedata.ResourceNone:
		return ""
	case gamedata.ResourceMineralRich:
		return "+Minerals"
	case gamedata.ResourceWaterWorld:
		return "+Water"
	case gamedata.ResourceRichFauna:
		return "+Fauna"
	case gamedata.ResourceRichSoil:
		return "+Soil"
	case gamedata.ResourceGoodClinic:
		return "+Good med"
	case gamedata.ResourceRobotWorkers:
		return "+Robots"
	case gamedata.ResourceDesert:
		return "-Desert"
	case gamedata.ResourcePoor:
		return "-Poor"
	case gamedata.ResourceLifeless:
		return "-Lifeless"
	case gamedata.ResourcePoorSoil:
		return "-Poor soil"
	case gamedata.ResourcePoorClinic:
		return "-Poor med"
	case gamedata.ResourceLackOfWorkers:
		return "-Low labor"
	case gamedata.ResourceIndustrial:
		return "~Industrial"
	}
	return r.String()
}

func eventOrNone(event string) string {
	if event == "" {
		return "none"
	}
	return event
}

var (
	resourceGreenStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	resourceRedStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	resourceYellowStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
)

func colorResource(r gamedata.Resource, text string) string {
	switch r {
	case gamedata.ResourceNone:
		return text
	case gamedata.ResourceMineralRich, gamedata.ResourceWaterWorld, gamedata.ResourceRichFauna,
		gamedata.ResourceRichSoil, gamedata.ResourceGoodClinic, gamedata.ResourceRobotWorkers:
		return resourceGreenStyle.Render(text)
	case gamedata.ResourceDesert, gamedata.ResourcePoor, gamedata.ResourceLifeless,
		gamedata.ResourcePoorSoil, gamedata.ResourcePoorClinic, gamedata.ResourceLackOfWorkers:
		return resourceRedStyle.Render(text)
	case gamedata.ResourceIndustrial:
		return resourceYellowStyle.Render(text)
	}
	return text
}

func resourceTradeHints(goods []gamedata.GoodDef, r gamedata.Resource) (cheap, expensive string) {
	resName := r.String()
	var cheapGoods, expensiveGoods []string
	for _, g := range goods {
		if g.CheapResource == resName {
			cheapGoods = append(cheapGoods, g.Name)
		}
		if g.ExpensiveResource == resName {
			expensiveGoods = append(expensiveGoods, g.Name)
		}
	}
	return strings.Join(cheapGoods, ", "), strings.Join(expensiveGoods, ", ")
}

func autoBookmarkNote(gs *game.GameState, sysIdx int) string {
	sys := gs.Data.Systems[sysIdx]
	sysState := gs.Systems[sysIdx]
	var parts []string
	if sysState.Event != "" {
		parts = append(parts, sysState.Event)
	}
	res := shortResource(sys.Resource)
	if res != "" {
		parts = append(parts, res)
	}
	parts = append(parts, shortTech(sys.TechLevel))
	return strings.Join(parts, ", ")
}
