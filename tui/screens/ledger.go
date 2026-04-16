package screens

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/the4ofus/spacetrader-tui/internal/game"
)

type ledgerSeries int

const (
	seriesCredits ledgerSeries = iota
	seriesNetWorth
	seriesBoth
)

type LedgerScreen struct {
	gs                *game.GameState
	series            ledgerSeries
	scroll            int
	scrollInitialized bool
}

func NewLedgerScreen(gs *game.GameState) *LedgerScreen {
	return &LedgerScreen{gs: gs, series: seriesBoth}
}

func (s *LedgerScreen) liveWorth() int {
	dp := &game.GameDataProvider{Data: s.gs.Data}
	return s.gs.Player.Worth(dp)
}

func (s *LedgerScreen) Init() tea.Cmd { return nil }

func (s *LedgerScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case msg.String() == "1":
			s.series = seriesCredits
		case msg.String() == "2":
			s.series = seriesNetWorth
		case msg.String() == "3":
			s.series = seriesBoth
		case msg.String() == "h" || msg.String() == "left":
			s.scroll -= 10
		case msg.String() == "l" || msg.String() == "right":
			s.scroll += 10
		case key.Matches(msg, Keys.Back):
			return s, func() tea.Msg { return NavigateMsg{Screen: ScreenSystem} }
		}
	}
	return s, nil
}

func (s *LedgerScreen) View() string {
	var b strings.Builder

	b.WriteString(HeaderStyle.Render("  PORTFOLIO  ") + "\n\n")

	snapshots := s.gs.Ledger
	if len(snapshots) < 2 {
		b.WriteString(DimStyle.Render("  Not enough data yet. Travel to more systems to build history.") + "\n")
		b.WriteString("\n" + DimStyle.Render("  esc back"))
		return b.String()
	}

	chartWidth := 65
	chartHeight := 16

	firstDay := snapshots[0].Day
	lastDay := snapshots[len(snapshots)-1].Day
	totalDays := lastDay - firstDay + 1
	maxScroll := totalDays - chartWidth
	if maxScroll < 0 {
		maxScroll = 0
	}
	if !s.scrollInitialized {
		s.scroll = maxScroll
		s.scrollInitialized = true
	}
	if s.scroll > maxScroll {
		s.scroll = maxScroll
	}
	if s.scroll < 0 {
		s.scroll = 0
	}
	startDay := firstDay + s.scroll

	credits, netWorth, valid := buildDayGrid(snapshots, startDay, chartWidth)

	_, maxVal := chartMinMax(credits, netWorth, valid, s.series)
	yAxisWidth := len(formatValue(maxVal)) + 1

	switch s.series {
	case seriesCredits:
		b.WriteString(CyanStyle.Render("  Credits") + "\n")
	case seriesNetWorth:
		b.WriteString(SuccessStyle.Render("  Net Worth") + "\n")
	case seriesBoth:
		b.WriteString(CyanStyle.Render("  Credits") + "  " + SuccessStyle.Render("  Net Worth") + "\n")
	}

	chart := renderChart(credits, netWorth, valid, chartWidth, chartHeight, s.series, yAxisWidth)
	b.WriteString(chart)

	canScrollLeft := s.scroll > 0
	canScrollRight := s.scroll < maxScroll
	xAxis := renderXAxis(startDay, chartWidth, yAxisWidth)
	indicator := "  "
	if canScrollLeft {
		indicator = "< "
	}
	indicatorRight := ""
	if canScrollRight {
		indicatorRight = " >"
	}
	axisLine := strings.TrimRight(xAxis, "\n")
	b.WriteString(indicator + axisLine[2:] + indicatorRight + "\n")
	b.WriteString(DimStyle.Render("  Chart updates on system arrival") + "\n")

	liveCredits := s.gs.Player.Credits
	liveWorth := s.liveWorth()

	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("  Credits:    %d cr\n", liveCredits))
	b.WriteString(fmt.Sprintf("  Net worth:  %d cr\n", liveWorth))
	credDelta := liveCredits - snapshots[0].Credits
	worthDelta := liveWorth - snapshots[0].NetWorth
	if credDelta >= 0 {
		b.WriteString(SuccessStyle.Render(fmt.Sprintf("  All-time:   +%d cr earned", credDelta)) + "\n")
	} else {
		b.WriteString(DangerStyle.Render(fmt.Sprintf("  All-time:   %d cr", credDelta)) + "\n")
	}
	if worthDelta >= 0 {
		b.WriteString(SuccessStyle.Render(fmt.Sprintf("  All-time:   +%d cr worth gained", worthDelta)) + "\n")
	} else {
		b.WriteString(DangerStyle.Render(fmt.Sprintf("  All-time:   %d cr worth", worthDelta)) + "\n")
	}

	b.WriteString("\n" + DimStyle.Render("  1 credits, 2 net worth, 3 both, h/l scroll, esc back"))
	return b.String()
}

func buildDayGrid(snapshots []game.DailySnapshot, startDay, width int) (credits, netWorth []int, valid []bool) {
	byDay := make(map[int]*game.DailySnapshot, len(snapshots))
	for i := range snapshots {
		byDay[snapshots[i].Day] = &snapshots[i]
	}
	credits = make([]int, width)
	netWorth = make([]int, width)
	valid = make([]bool, width)
	for col := 0; col < width; col++ {
		day := startDay + col
		if snap, ok := byDay[day]; ok {
			credits[col] = snap.Credits
			netWorth[col] = snap.NetWorth
			valid[col] = true
		}
	}
	return
}

func chartMinMax(credits, netWorth []int, valid []bool, series ledgerSeries) (int, int) {
	minVal := 0
	maxVal := 0
	for i, ok := range valid {
		if !ok {
			continue
		}
		switch series {
		case seriesCredits:
			if credits[i] > maxVal {
				maxVal = credits[i]
			}
		case seriesNetWorth:
			if netWorth[i] > maxVal {
				maxVal = netWorth[i]
			}
		case seriesBoth:
			if credits[i] > maxVal {
				maxVal = credits[i]
			}
			if netWorth[i] > maxVal {
				maxVal = netWorth[i]
			}
		}
	}
	if maxVal == minVal {
		maxVal = minVal + 1
	}
	return minVal, maxVal
}

func renderChart(credits, netWorth []int, valid []bool, width, height int, series ledgerSeries, yAxisWidth int) string {
	minVal, maxVal := chartMinMax(credits, netWorth, valid, series)
	dotRows := height * 4
	dotCols := width * 2

	grid1 := make([][]bool, dotRows)
	grid2 := make([][]bool, dotRows)
	for i := range grid1 {
		grid1[i] = make([]bool, dotCols)
		grid2[i] = make([]bool, dotCols)
	}

	plotPoint := func(grid [][]bool, col, val int) {
		y := (val - minVal) * (dotRows - 1) / (maxVal - minVal)
		if y < 0 {
			y = 0
		}
		if y >= dotRows {
			y = dotRows - 1
		}
		grid[y][col*2] = true
		grid[y][col*2+1] = true
	}

	for col, ok := range valid {
		if !ok {
			continue
		}
		switch series {
		case seriesCredits:
			plotPoint(grid1, col, credits[col])
		case seriesNetWorth:
			plotPoint(grid1, col, netWorth[col])
		case seriesBoth:
			plotPoint(grid1, col, credits[col])
			plotPoint(grid2, col, netWorth[col])
		}
	}

	var b strings.Builder
	for row := height - 1; row >= 0; row-- {
		label := ""
		if row == height-1 {
			label = formatValue(maxVal)
		} else if row == 0 {
			label = formatValue(minVal)
		} else if row == height/2 {
			label = formatValue((maxVal + minVal) / 2)
		}
		b.WriteString(fmt.Sprintf("  %*s|", yAxisWidth, label))

		for col := 0; col < width; col++ {
			dotR0 := row * 4
			var pattern byte
			has1 := false
			has2 := false
			for dr := 0; dr < 4; dr++ {
				dotRow := dotR0 + dr
				if dotRow >= dotRows {
					continue
				}
				gridRow := dotRow
				for dc := 0; dc < 2; dc++ {
					dotCol := col*2 + dc
					if dotCol >= dotCols {
						continue
					}
					if grid1[gridRow][dotCol] {
						pattern |= brailleBit(dr, dc)
						has1 = true
					}
					if grid2[gridRow][dotCol] {
						pattern |= brailleBit(dr, dc)
						has2 = true
					}
				}
			}

			if pattern == 0 {
				b.WriteString(" ")
			} else if series == seriesBoth {
				ch := string(brailleChar(pattern))
				if has1 && has2 {
					b.WriteString(SelectedStyle.Render(ch))
				} else if has1 {
					b.WriteString(CyanStyle.Render(ch))
				} else {
					b.WriteString(SuccessStyle.Render(ch))
				}
			} else if series == seriesCredits {
				b.WriteString(CyanStyle.Render(string(brailleChar(pattern))))
			} else {
				b.WriteString(SuccessStyle.Render(string(brailleChar(pattern))))
			}
		}
		b.WriteString("\n")
	}
	return b.String()
}

func renderXAxis(startDay, width, yAxisWidth int) string {
	buf := make([]byte, width)
	for i := range buf {
		buf[i] = ' '
	}
	rightmost := 0
	for col := 0; col < width; col++ {
		day := startDay + col
		if day > 0 && day%10 == 0 {
			label := fmt.Sprintf("%d", day)
			if col >= rightmost && col+len(label) <= width {
				copy(buf[col:], label)
				rightmost = col + len(label) + 1
			}
		}
	}
	prefix := strings.Repeat(" ", yAxisWidth+3)
	return prefix + string(buf) + "\n"
}

func brailleBit(row, col int) byte {
	if col == 0 {
		switch row {
		case 0:
			return 0x01
		case 1:
			return 0x02
		case 2:
			return 0x04
		case 3:
			return 0x40
		}
	} else {
		switch row {
		case 0:
			return 0x08
		case 1:
			return 0x10
		case 2:
			return 0x20
		case 3:
			return 0x80
		}
	}
	return 0
}

func brailleChar(pattern byte) rune {
	return rune(0x2800 + int(pattern))
}

func formatValue(v int) string {
	if v >= 1000000 {
		return fmt.Sprintf("%.1fM", float64(v)/1000000)
	}
	if v >= 1000 {
		return fmt.Sprintf("%.0fk", float64(v)/1000)
	}
	return fmt.Sprintf("%d", v)
}
