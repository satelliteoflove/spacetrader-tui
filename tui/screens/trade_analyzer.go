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

const maxAnalyzerRows = 20

type analyzerRow struct {
	goodName string
	fromIdx  int
	toIdx    int
	fromName string
	toName   string
	buy      int
	sell     int
	profit   int
	dist     float64
	stale    bool
}

type TradeAnalyzerScreen struct {
	gs   *game.GameState
	rows []analyzerRow
}

func NewTradeAnalyzerScreen(gs *game.GameState) *TradeAnalyzerScreen {
	s := &TradeAnalyzerScreen{gs: gs}
	s.rows = computeAnalyzerRows(gs)
	return s
}

func computeAnalyzerRows(gs *game.GameState) []analyzerRow {
	if gs.TradeInfo == nil {
		return nil
	}
	curIdx := gs.CurrentSystemID
	cur := gs.Data.Systems[curIdx]
	maxRange := float64(gs.EffectiveRange())

	var rows []analyzerRow
	for targetIdx := range gs.TradeInfo {
		if targetIdx == curIdx {
			continue
		}
		if targetIdx < 0 || targetIdx >= len(gs.Data.Systems) {
			continue
		}
		target := gs.Data.Systems[targetIdx]
		dist := formula.Distance(cur.X, cur.Y, target.X, target.Y)
		if dist > maxRange {
			continue
		}
		stale, _ := gs.IsTradeInfoStale(targetIdx)

		for g, good := range gs.Data.Goods {
			curBuy := game.BuyPriceAt(gs, curIdx, g)
			curSell := game.SellPriceAt(gs, curIdx, g)
			tgtBuy := game.BuyPriceAt(gs, targetIdx, g)
			tgtSell := game.SellPriceAt(gs, targetIdx, g)

			if curBuy > 0 && tgtSell > 0 {
				profit := tgtSell - curBuy
				if profit > 0 {
					rows = append(rows, analyzerRow{
						goodName: good.Name,
						fromIdx:  curIdx,
						toIdx:    targetIdx,
						fromName: cur.Name,
						toName:   target.Name,
						buy:      curBuy,
						sell:     tgtSell,
						profit:   profit,
						dist:     dist,
						stale:    stale,
					})
				}
			}
			if tgtBuy > 0 && curSell > 0 {
				profit := curSell - tgtBuy
				if profit > 0 {
					rows = append(rows, analyzerRow{
						goodName: good.Name,
						fromIdx:  targetIdx,
						toIdx:    curIdx,
						fromName: target.Name,
						toName:   cur.Name,
						buy:      tgtBuy,
						sell:     curSell,
						profit:   profit,
						dist:     dist,
						stale:    stale,
					})
				}
			}
		}
	}

	sort.Slice(rows, func(i, j int) bool {
		return rows[i].profit > rows[j].profit
	})
	if len(rows) > maxAnalyzerRows {
		rows = rows[:maxAnalyzerRows]
	}
	return rows
}

func (s *TradeAnalyzerScreen) Init() tea.Cmd { return nil }

func (s *TradeAnalyzerScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if key.Matches(msg, Keys.Back) {
			return s, func() tea.Msg { return NavigateMsg{Screen: ScreenMarket} }
		}
	}
	return s, nil
}

func (s *TradeAnalyzerScreen) View() string {
	var b strings.Builder

	b.WriteString(HeaderStyle.Render("  TRADE ANALYZER  ") + "\n")
	b.WriteString(DimStyle.Render(fmt.Sprintf("  Routes from known systems within %d parsecs.", gs_effectiveRange(s.gs))) + "\n\n")

	if len(s.rows) == 0 {
		b.WriteString(DimStyle.Render("  No profitable routes in known in-range systems.") + "\n")
		b.WriteString(DimStyle.Render("  Visit more systems or buy trade info to widen coverage.") + "\n")
		b.WriteString("\n" + DimStyle.Render("  esc back"))
		return b.String()
	}

	header := fmt.Sprintf("  %-14s %-22s %6s %6s %7s %5s %s",
		"GOOD", "ROUTE", "BUY", "SELL", "PROFIT", "DIST", "")
	b.WriteString(DimStyle.Render(header) + "\n")
	b.WriteString("  " + strings.Repeat("-", 70) + "\n")

	for _, r := range s.rows {
		route := fmt.Sprintf("%s -> %s", shortSysName(r.fromName), shortSysName(r.toName))
		if len(route) > 22 {
			route = route[:22]
		}
		line := fmt.Sprintf("  %-14s %-22s %6d %6d %7d %5.1f",
			truncateName(r.goodName, 14), route, r.buy, r.sell, r.profit, r.dist)
		tag := ""
		if r.stale {
			tag = " stale"
		}
		if r.stale {
			b.WriteString(DimStyle.Render(line+tag) + "\n")
		} else {
			b.WriteString(line + "\n")
		}
	}

	b.WriteString("\n" + DimStyle.Render("  esc back"))
	return b.String()
}

func gs_effectiveRange(gs *game.GameState) int {
	return int(math.Ceil(float64(gs.EffectiveRange())))
}

func shortSysName(name string) string {
	if len(name) <= 10 {
		return name
	}
	return name[:10]
}

func truncateName(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}

func (s *TradeAnalyzerScreen) HelpTitle() string { return "Trade Analyzer" }

func (s *TradeAnalyzerScreen) HelpGroups() []KeyGroup {
	return []KeyGroup{
		{
			Title: "About",
			Bindings: []KeyBinding{
				{Keys: "", Desc: "Shows top trades to/from in-range systems"},
				{Keys: "", Desc: "you have visited or purchased info for."},
				{Keys: "", Desc: "Stale rows (>5 days) are dimmed."},
			},
		},
	}
}
