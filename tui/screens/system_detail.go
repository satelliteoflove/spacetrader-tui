package screens

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/the4ofus/spacetrader-tui/internal/formula"
	"github.com/the4ofus/spacetrader-tui/internal/game"
	"github.com/the4ofus/spacetrader-tui/internal/market"
)

type SystemDetailScreen struct {
	gs           *game.GameState
	sysIdx       int
	returnScreen ScreenType
	message      string
}

func NewSystemDetailScreen(gs *game.GameState, sysIdx int, returnScreen ScreenType) *SystemDetailScreen {
	return &SystemDetailScreen{gs: gs, sysIdx: sysIdx, returnScreen: returnScreen}
}

func (s *SystemDetailScreen) Init() tea.Cmd { return nil }

func (s *SystemDetailScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case msg.String() == "i":
			s.message = buyTradeInfo(s.gs, s.sysIdx)
		case key.Matches(msg, Keys.Back):
			sysIdx := s.sysIdx
			ret := s.returnScreen
			return s, func() tea.Msg {
				return NavigateMsg{Screen: ret, SelectedSystem: sysIdx}
			}
		}
	}
	return s, nil
}

func (s *SystemDetailScreen) View() string {
	var b strings.Builder

	sys := s.gs.Data.Systems[s.sysIdx]
	sysState := s.gs.Systems[s.sysIdx]
	cur := s.gs.Data.Systems[s.gs.CurrentSystemID]
	dist := formula.Distance(cur.X, cur.Y, sys.X, sys.Y)

	b.WriteString(HeaderStyle.Render(fmt.Sprintf("  %s  ", sys.Name)) + "\n\n")

	b.WriteString(fmt.Sprintf("  Tech Level: %s\n", sys.TechLevel))
	b.WriteString(fmt.Sprintf("  Government: %s\n", sys.PoliticalSystem))
	b.WriteString(fmt.Sprintf("  Size: %s\n", sys.Size))
	if sys.Resource.String() != "No Special Resources" {
		b.WriteString(fmt.Sprintf("  Resource: %s\n", sys.Resource))
	}
	b.WriteString(fmt.Sprintf("  Distance: %.1f parsecs\n", dist))

	if sysState.Event != "" {
		b.WriteString(DangerStyle.Render(fmt.Sprintf("  Event: %s", sysState.Event)) + "\n")
	}

	for _, wh := range s.gs.Wormholes {
		if wh.SystemA == s.sysIdx {
			b.WriteString(SuccessStyle.Render(fmt.Sprintf("  Wormhole to %s", s.gs.Data.Systems[wh.SystemB].Name)) + "\n")
		} else if wh.SystemB == s.sysIdx {
			b.WriteString(SuccessStyle.Render(fmt.Sprintf("  Wormhole to %s", s.gs.Data.Systems[wh.SystemA].Name)) + "\n")
		}
	}

	if bm, ok := s.gs.GetBookmark(s.sysIdx); ok && bm.Note != "" {
		b.WriteString(SelectedStyle.Render(fmt.Sprintf("  Bookmarked: %s", bm.Note)) + "\n")
	}

	b.WriteString("\n")

	info, hasInfo := s.gs.GetTradeInfo(s.sysIdx)
	stale := false
	age := 0
	if hasInfo {
		stale, age = s.gs.IsTradeInfoStale(s.sysIdx)
	}

	if hasInfo {
		header := "  Trade Prices"
		if stale {
			if age == 1 {
				header += " (stale -- 1 day old)"
			} else {
				header += fmt.Sprintf(" (stale -- %d days old)", age)
			}
			b.WriteString(DimStyle.Render(header) + "\n")
		} else {
			if age == 0 {
				header += " (today)"
			} else if age == 1 {
				header += " (1 day old)"
			} else {
				header += fmt.Sprintf(" (%d days old)", age)
			}
			b.WriteString(CyanStyle.Render(header) + "\n")
		}

		b.WriteString(DimStyle.Render(fmt.Sprintf("  %-12s %8s %8s %8s", "GOOD", "EST", "AVG", "CARGO")) + "\n")

		for g, good := range s.gs.Data.Goods {
			price := info.Prices[g]
			if price <= 0 {
				continue
			}
			avg := market.AveragePrice(good, s.gs.Data.Systems)

			priceStr := fmt.Sprintf("~%d", price)
			avgStr := fmt.Sprintf("%d", avg)

			held := s.gs.Player.Cargo[g]
			cargoStr := ""
			if held > 0 {
				cargoStr = fmt.Sprintf("%d", held)
			}

			line := fmt.Sprintf("  %-12s %8s %8s %8s", good.Name, priceStr, avgStr, cargoStr)

			if stale {
				b.WriteString(DimStyle.Render(line))
			} else {
				pct := 0
				if avg > 0 {
					pct = (price - avg) * 100 / avg
				}
				if held > 0 {
					costPer := 0
					if s.gs.Player.CargoCost[g] > 0 {
						costPer = s.gs.Player.CargoCost[g] / held
					}
					profit := price - costPer
					if costPer > 0 && profit > 0 {
						b.WriteString(line + SuccessStyle.Render(fmt.Sprintf("  +%d/ea", profit)))
					} else if costPer > 0 && profit < 0 {
						b.WriteString(line + DangerStyle.Render(fmt.Sprintf("  %d/ea", profit)))
					} else {
						b.WriteString(line)
					}
				} else if pct <= -5 {
					b.WriteString(SuccessStyle.Render(line))
				} else if pct >= 5 {
					b.WriteString(DangerStyle.Render(line))
				} else {
					b.WriteString(line)
				}
			}
			b.WriteString("\n")
		}
	} else {
		b.WriteString(DimStyle.Render("  No trade info available") + "\n")
		distToSys := formula.Distance(cur.X, cur.Y, sys.X, sys.Y)
		if distToSys <= game.TradeInfoMaxRange {
			b.WriteString(DimStyle.Render(fmt.Sprintf("  Press i to purchase trade info (%d cr)", game.TradeInfoBuyCost)) + "\n")
		} else {
			b.WriteString(DimStyle.Render(fmt.Sprintf("  Too far to purchase info (%.1f parsecs, max %.0f)", distToSys, game.TradeInfoMaxRange)) + "\n")
		}
	}

	if s.message != "" {
		b.WriteString("\n  " + s.message + "\n")
	}

	b.WriteString("\n" + DimStyle.Render("  i buy trade info, esc back"))

	return b.String()
}
