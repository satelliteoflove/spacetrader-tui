package screens

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/the4ofus/spacetrader-tui/internal/game"
	"github.com/the4ofus/spacetrader-tui/internal/shipyard"
	"github.com/the4ofus/spacetrader-tui/internal/travel"
)

type RoutePlannerScreen struct {
	gs           *game.GameState
	route        travel.Route
	trades       []travel.HopTradeInfo
	cursor       int
	currentHop   int
	destIdx      int
	confirming   bool
	message      string
	isActive     bool
	returnScreen ScreenType
}

func NewRoutePlannerScreen(gs *game.GameState, destIdx int, returnScreen ScreenType) *RoutePlannerScreen {
	isActive := gs.HasActiveRoute && gs.ActiveRoute == destIdx

	var originIdx int
	if isActive {
		originIdx = gs.ActiveRouteOrigin
	} else {
		originIdx = gs.CurrentSystemID
	}

	route := travel.FindRouteFrom(gs, originIdx, destIdx)
	var trades []travel.HopTradeInfo
	if route.Reachable {
		trades = travel.AnalyzeRouteTrades(gs, route)
	}

	currentHop := 0
	for i, hop := range route.Hops {
		if hop.SystemIdx == gs.CurrentSystemID {
			currentHop = i
		}
	}

	return &RoutePlannerScreen{
		gs:           gs,
		route:        route,
		trades:       trades,
		cursor:       currentHop,
		currentHop:   currentHop,
		destIdx:      destIdx,
		isActive:     isActive,
		returnScreen: returnScreen,
	}
}

func (s *RoutePlannerScreen) Init() tea.Cmd { return nil }

func (s *RoutePlannerScreen) setRoute() {
	s.gs.HasActiveRoute = true
	s.gs.ActiveRoute = s.destIdx
	s.gs.ActiveRouteOrigin = s.gs.CurrentSystemID
	s.isActive = true
}

func (s *RoutePlannerScreen) nextHopIdx() int {
	if s.currentHop+1 < len(s.route.Hops) {
		return s.currentHop + 1
	}
	return -1
}

func (s *RoutePlannerScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if s.confirming {
		if msg, ok := msg.(tea.KeyMsg); ok {
			switch msg.String() {
			case "y":
				s.confirming = false
				next := s.nextHopIdx()
				if next < 0 {
					return s, nil
				}
				nextHop := s.route.Hops[next]
				s.setRoute()
				if nextHop.IsWormhole {
					ok, wmsg := game.TravelWormhole(s.gs)
					if !ok {
						s.message = wmsg
						return s, nil
					}
					return s, func() tea.Msg { return TravelMsg{DestIdx: s.gs.CurrentSystemID} }
				}
				destIdx := nextHop.SystemIdx
				result := travel.ExecuteTravel(s.gs, destIdx)
				if !result.Success {
					s.message = result.Message
					return s, nil
				}
				return s, func() tea.Msg { return TravelMsg{DestIdx: destIdx} }
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
			if len(s.route.Hops) > 0 {
				s.cursor = wrapCursor(s.cursor, -1, len(s.route.Hops))
			}
		case key.Matches(msg, Keys.Down):
			if len(s.route.Hops) > 0 {
				s.cursor = wrapCursor(s.cursor, 1, len(s.route.Hops))
			}
		case msg.String() == "t":
			return s, func() tea.Msg { return NavigateMsg{Screen: ScreenMarket} }
		case msg.String() == "m":
			return s, func() tea.Msg { return NavigateMsg{Screen: ScreenGalacticChart, SelectedSystem: s.destIdx} }
		case msg.String() == "r":
			if s.route.Reachable && !s.isActive {
				s.setRoute()
				s.message = SuccessStyle.Render("Route set!")
			} else if s.isActive {
				s.gs.HasActiveRoute = false
				s.isActive = false
				s.message = DimStyle.Render("Route cleared.")
			}
		case msg.String() == "f":
			result := shipyard.Refuel(s.gs)
			s.message = result.Message
		case msg.String() == "b":
			if s.route.Reachable && s.cursor < len(s.route.Hops) {
				sysIdx := s.route.Hops[s.cursor].SystemIdx
				s.gs.ToggleBookmark(sysIdx, "")
			}
		case key.Matches(msg, Keys.Enter):
			next := s.nextHopIdx()
			if s.route.Reachable && next >= 0 {
				nextHop := s.route.Hops[next]
				name := s.gs.Data.Systems[nextHop.SystemIdx].Name
				if nextHop.IsWormhole {
					fee := game.WormholeTax(s.gs)
					if s.gs.Player.Credits < fee {
						s.message = DangerStyle.Render(fmt.Sprintf("Need %d cr for wormhole fee", fee))
					} else {
						s.message = SelectedStyle.Render(fmt.Sprintf("Wormhole to %s? (%d cr fee) (y/n)", name, fee))
						s.confirming = true
					}
				} else {
					if nextHop.FuelCost > s.gs.Player.Ship.Fuel {
						s.message = DangerStyle.Render(fmt.Sprintf("Need %d fuel, have %d -- refuel first", nextHop.FuelCost, s.gs.Player.Ship.Fuel))
					} else {
						s.message = SelectedStyle.Render(fmt.Sprintf("Travel to %s? (y/n)", name))
						s.confirming = true
					}
				}
			}
		case key.Matches(msg, Keys.Back):
			ret := s.returnScreen
			destIdx := s.destIdx
			return s, func() tea.Msg { return NavigateMsg{Screen: ret, SelectedSystem: destIdx} }
		}
	}
	return s, nil
}

func (s *RoutePlannerScreen) View() string {
	var b strings.Builder

	b.WriteString(HeaderStyle.Render("  ROUTE PLANNER  ") + "\n")

	if !s.route.Reachable {
		destName := s.gs.Data.Systems[s.destIdx].Name
		b.WriteString(fmt.Sprintf("  To: %s\n\n", destName))
		b.WriteString("  No route found. The destination is beyond reach\n")
		b.WriteString("  even with refueling at every stop.\n\n")
		b.WriteString(DimStyle.Render("  Consider upgrading your ship for greater range,") + "\n")
		b.WriteString(DimStyle.Render("  or check for wormhole connections.") + "\n")
		b.WriteString("\n" + DimStyle.Render("  esc back"))
		return b.String()
	}

	originName := s.gs.Data.Systems[s.route.Hops[0].SystemIdx].Name
	destName := s.gs.Data.Systems[s.destIdx].Name
	totalHops := len(s.route.Hops) - 1
	remainHops := totalHops - s.currentHop

	titleLine := fmt.Sprintf("  %s -> %s  (%d/%d hops)", originName, destName, remainHops, totalHops)
	if s.isActive {
		titleLine += "  " + SuccessStyle.Render("[ACTIVE]")
	}
	b.WriteString(titleLine + "\n")

	costParts := []string{fmt.Sprintf("Refuel: ~%d cr", s.route.TotalRefuel)}
	if s.route.TotalWormhole > 0 {
		costParts = append(costParts, fmt.Sprintf("WH fees: %d cr", s.route.TotalWormhole))
	}
	b.WriteString(DimStyle.Render("  "+strings.Join(costParts, "  |  ")) + "\n\n")

	header := fmt.Sprintf("  %-3s %-16s %5s %5s %8s  %s", "#", "SYSTEM", "DIST", "FUEL", "REFUEL", "")
	b.WriteString(DimStyle.Render(header) + "\n")
	b.WriteString("  " + strings.Repeat("-", 52) + "\n")

	for i, hop := range s.route.Hops {
		sysName := s.gs.Data.Systems[hop.SystemIdx].Name
		isPast := i < s.currentHop
		isCurrent := i == s.currentHop

		var distStr, fuelStr, refuelStr, note string

		if isCurrent {
			distStr = "(here)"
			fuelStr = "--"
			refuelStr = "--"
		} else if i == 0 && isPast {
			distStr = "start"
			fuelStr = "--"
			refuelStr = "--"
		} else if hop.IsWormhole {
			distStr = "WH"
			fuelStr = "0"
			refuelStr = fmt.Sprintf("%d cr", hop.WormholeFee)
		} else {
			distStr = fmt.Sprintf("%.0f", hop.Distance)
			fuelStr = fmt.Sprintf("%d", hop.FuelCost)
			refuelStr = fmt.Sprintf("%d cr", hop.RefuelCost)
		}

		if i == len(s.route.Hops)-1 {
			note = "DEST"
		}

		line := fmt.Sprintf("%-3d %-16s %5s %5s %8s  %s",
			i+1, truncate(sysName, 16), distStr, fuelStr, refuelStr, note)

		if i == s.cursor {
			b.WriteString(SelectedStyle.Render("> ") + SelectedStyle.Render(line) + "\n")
		} else if isPast {
			b.WriteString(DimStyle.Render("  "+line) + "\n")
		} else {
			b.WriteString("  " + line + "\n")
		}
	}

	if s.cursor < len(s.trades) {
		b.WriteString("\n")
		info := s.trades[s.cursor]
		fromName := truncate(s.gs.Data.Systems[info.FromIdx].Name, 12)
		toName := truncate(s.gs.Data.Systems[info.ToIdx].Name, 12)
		b.WriteString(DimStyle.Render(fmt.Sprintf("  -- %s -> %s --", fromName, toName)) + "\n")

		if info.NoFromInfo || info.NoToInfo {
			missing := ""
			if info.NoFromInfo && info.NoToInfo {
				missing = "No trade info for either system."
			} else if info.NoFromInfo {
				missing = fmt.Sprintf("No trade info for %s.", truncate(s.gs.Data.Systems[info.FromIdx].Name, 16))
			} else {
				missing = fmt.Sprintf("No trade info for %s.", truncate(s.gs.Data.Systems[info.ToIdx].Name, 16))
			}
			b.WriteString(DimStyle.Render("  "+missing) + "\n")
			b.WriteString(DimStyle.Render("  Visit or purchase info (i) from Navigation.") + "\n")
		} else if len(info.Trades) == 0 {
			b.WriteString(DimStyle.Render("  No profitable trades for this hop.") + "\n")
		} else {
			staleNote := ""
			if info.FromStale || info.ToStale {
				staleNote = " (stale info -- estimates may be off)"
			}
			b.WriteString(DimStyle.Render(fmt.Sprintf("  %-12s %6s %6s %8s", "GOOD", "~BUY", "~SELL", "~PROFIT")) + "\n")
			for _, t := range info.Trades {
				profitStr := SuccessStyle.Render(fmt.Sprintf("~+%d/unit", t.Profit))
				b.WriteString(fmt.Sprintf("  %-12s %6d %6d  %s\n", t.GoodName, t.BuyPrice, t.SellPrice, profitStr))
			}
			if staleNote != "" {
				b.WriteString(DangerStyle.Render("  "+staleNote) + "\n")
			} else {
				b.WriteString(DimStyle.Render("  prices may change on arrival") + "\n")
			}
		}
	}

	if s.message != "" {
		b.WriteString("\n  " + s.message + "\n")
	} else {
		next := s.nextHopIdx()
		if next >= 0 {
			nextHop := s.route.Hops[next]
			name := s.gs.Data.Systems[nextHop.SystemIdx].Name
			if nextHop.IsWormhole {
				b.WriteString(fmt.Sprintf("\n  Next: %s via wormhole (%d cr fee)\n", name, game.WormholeTax(s.gs)))
			} else {
				b.WriteString(fmt.Sprintf("\n  Next: %s (%d fuel, you have %d)\n", name, nextHop.FuelCost, s.gs.Player.Ship.Fuel))
			}
		} else {
			b.WriteString("\n  " + SuccessStyle.Render("You are at the destination!") + "\n")
		}
	}

	helpLine1 := "  j/k navigate  t trade  m map  f refuel  enter travel"
	helpLine2 := "  "
	if s.isActive {
		helpLine2 += "r clear route"
	} else {
		helpLine2 += "r set route"
	}
	helpLine2 += "  esc back"
	b.WriteString("\n" + DimStyle.Render(helpLine1))
	b.WriteString("\n" + DimStyle.Render(helpLine2))
	return b.String()
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-1] + "."
}
