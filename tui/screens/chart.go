package screens

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/the4ofus/spacetrader-tui/internal/game"
	"github.com/the4ofus/spacetrader-tui/internal/gamedata"
	"github.com/the4ofus/spacetrader-tui/internal/shipyard"
	"github.com/the4ofus/spacetrader-tui/internal/travel"
)

type ChartScreen struct {
	gs      *game.GameState
	cursor  int
	systems []travel.ReachableSystem
	message string
}

func NewChartScreen(gs *game.GameState) *ChartScreen {
	return &ChartScreen{
		gs:      gs,
		systems: travel.ReachableSystems(gs),
	}
}

func (s *ChartScreen) Init() tea.Cmd { return nil }

func (s *ChartScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, Keys.Up):
			s.cursor = wrapCursor(s.cursor, -1, len(s.systems))
		case key.Matches(msg, Keys.Down):
			s.cursor = wrapCursor(s.cursor, 1, len(s.systems))
		case key.Matches(msg, Keys.Enter):
			if len(s.systems) == 0 {
				return s, nil
			}
			dest := s.systems[s.cursor]
			result := travel.ExecuteTravel(s.gs, dest.Index)
			if !result.Success {
				s.message = result.Message
				return s, nil
			}
			s.message = result.Message
			return s, func() tea.Msg { return TravelMsg{DestIdx: dest.Index} }
		case msg.String() == "r":
			result := shipyard.Refuel(s.gs)
			s.message = result.Message
			s.systems = travel.ReachableSystems(s.gs)
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

	b.WriteString("\n")

	if len(s.systems) == 0 {
		b.WriteString("  No systems in range.\n")
		if s.gs.Player.Ship.Fuel < shipDef.Range {
			b.WriteString("  " + DimStyle.Render("Press r to refuel.") + "\n")
		}
	} else {
		b.WriteString(fmt.Sprintf("  %-16s %4s  %-15s %-16s %s\n",
			"SYSTEM", "DIST", "TECH", "GOVERNMENT", "RESOURCE"))
		b.WriteString("  " + strings.Repeat("-", 68) + "\n")

		for i, sys := range s.systems {
			sysDef := s.gs.Data.Systems[sys.Index]
			visited := s.gs.Systems[sys.Index].Visited

			visitMark := " "
			if visited {
				visitMark = "*"
			}

			techStr := shortTech(sysDef.TechLevel)
			resStr := shortResource(sysDef.Resource)

			line := fmt.Sprintf("%-16s %4.1f  %-15s %-16s %s %s",
				sys.Name, sys.Distance, techStr,
				sysDef.PoliticalSystem, resStr, visitMark)

			if i == s.cursor {
				b.WriteString(SelectedStyle.Render("> ") + line + "\n")
			} else {
				b.WriteString("  " + line + "\n")
			}
		}

		b.WriteString("\n" + DimStyle.Render("  * = visited"))
	}

	if s.message != "" {
		b.WriteString("\n  " + s.message)
	}

	if len(s.systems) > 0 && s.cursor < len(s.systems) {
		b.WriteString("\n")
		destIdx := s.systems[s.cursor].Index
		destDef := s.gs.Data.Systems[destIdx]
		destState := s.gs.Systems[destIdx]

		b.WriteString(DimStyle.Render(fmt.Sprintf("  --- %s ---", destDef.Name)) + "\n")

		availableGoods := 0
		for g := 0; g < game.NumGoods; g++ {
			if destState.Prices[g] > 0 {
				availableGoods++
			}
		}
		if s.gs.Systems[destIdx].Visited {
			b.WriteString(DimStyle.Render(fmt.Sprintf("  %d goods available  |  Event: %s",
				availableGoods, eventOrNone(destState.Event))) + "\n")
		} else {
			b.WriteString(DimStyle.Render("  Not yet visited") + "\n")
		}
	}

	b.WriteString("\n" + DimStyle.Render("  enter travel, r refuel, w wormhole, esc back"))
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
		return "Minerals"
	case gamedata.ResourceWaterWorld:
		return "Water"
	case gamedata.ResourceRichFauna:
		return "Fauna"
	case gamedata.ResourceRichSoil:
		return "Soil"
	case gamedata.ResourcePoorSoil:
		return "Poor soil"
	case gamedata.ResourcePoorClinic:
		return "Poor med"
	case gamedata.ResourceGoodClinic:
		return "Good med"
	case gamedata.ResourceLackOfWorkers:
		return "Low labor"
	case gamedata.ResourceRobotWorkers:
		return "Robots"
	}
	return r.String()
}

func eventOrNone(event string) string {
	if event == "" {
		return "none"
	}
	return event
}
