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
	gs     *game.GameState
	series ledgerSeries
}

func NewLedgerScreen(gs *game.GameState) *LedgerScreen {
	return &LedgerScreen{gs: gs, series: seriesBoth}
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
		case key.Matches(msg, Keys.Back):
			return s, func() tea.Msg { return NavigateMsg{Screen: ScreenSystem} }
		}
	}
	return s, nil
}

func (s *LedgerScreen) View() string {
	var b strings.Builder

	b.WriteString(HeaderStyle.Render("  LEDGER  ") + "\n\n")

	snapshots := s.gs.Ledger
	if len(snapshots) < 2 {
		b.WriteString(DimStyle.Render("  Not enough data yet. Travel to more systems to build history.") + "\n")
		b.WriteString("\n" + DimStyle.Render("  esc back"))
		return b.String()
	}

	chartWidth := 60
	chartHeight := 16

	var credits, netWorth []int
	var days []int
	for _, snap := range snapshots {
		credits = append(credits, snap.Credits)
		netWorth = append(netWorth, snap.NetWorth)
		days = append(days, snap.Day)
	}

	switch s.series {
	case seriesCredits:
		b.WriteString(CyanStyle.Render("  Credits") + "\n")
		chart := renderBrailleChart(credits, chartWidth, chartHeight)
		b.WriteString(chart)
	case seriesNetWorth:
		b.WriteString(SuccessStyle.Render("  Net Worth") + "\n")
		chart := renderBrailleChart(netWorth, chartWidth, chartHeight)
		b.WriteString(chart)
	case seriesBoth:
		b.WriteString(CyanStyle.Render("  Credits") + "  " + SuccessStyle.Render("  Net Worth") + "\n")
		chart := renderBrailleDualChart(credits, netWorth, chartWidth, chartHeight)
		b.WriteString(chart)
	}

	first := snapshots[0]
	last := snapshots[len(snapshots)-1]
	b.WriteString(fmt.Sprintf("  Day %-6d", first.Day))
	dayLabel := fmt.Sprintf("Day %d", last.Day)
	pad := chartWidth*2 - 8 - len(dayLabel)
	if pad < 1 {
		pad = 1
	}
	b.WriteString(strings.Repeat(" ", pad) + dayLabel + "\n")

	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("  Current credits:   %d cr\n", last.Credits))
	b.WriteString(fmt.Sprintf("  Current net worth: %d cr\n", last.NetWorth))
	credDelta := last.Credits - first.Credits
	worthDelta := last.NetWorth - first.NetWorth
	if credDelta >= 0 {
		b.WriteString(SuccessStyle.Render(fmt.Sprintf("  Credit change:     +%d cr", credDelta)) + "\n")
	} else {
		b.WriteString(DangerStyle.Render(fmt.Sprintf("  Credit change:     %d cr", credDelta)) + "\n")
	}
	if worthDelta >= 0 {
		b.WriteString(SuccessStyle.Render(fmt.Sprintf("  Worth change:      +%d cr", worthDelta)) + "\n")
	} else {
		b.WriteString(DangerStyle.Render(fmt.Sprintf("  Worth change:      %d cr", worthDelta)) + "\n")
	}

	b.WriteString("\n" + DimStyle.Render("  1 credits, 2 net worth, 3 both, esc back"))
	return b.String()
}

func renderBrailleChart(data []int, width, height int) string {
	if len(data) == 0 {
		return ""
	}

	samples := resample(data, width*2)

	minVal, maxVal := samples[0], samples[0]
	for _, v := range samples {
		if v < minVal {
			minVal = v
		}
		if v > maxVal {
			maxVal = v
		}
	}
	if minVal > 0 {
		minVal = 0
	}
	if maxVal == minVal {
		maxVal = minVal + 1
	}

	dotRows := height * 4
	grid := make([][]bool, dotRows)
	for i := range grid {
		grid[i] = make([]bool, width*2)
	}

	for x, val := range samples {
		y := (val - minVal) * (dotRows - 1) / (maxVal - minVal)
		if y < 0 {
			y = 0
		}
		if y >= dotRows {
			y = dotRows - 1
		}
		for fill := 0; fill <= y; fill++ {
			grid[fill][x] = true
		}
	}

	return renderBrailleGrid(grid, width, height, minVal, maxVal, CyanStyle.Render)
}

func renderBrailleDualChart(data1, data2 []int, width, height int) string {
	if len(data1) == 0 || len(data2) == 0 {
		return ""
	}

	samples1 := resample(data1, width*2)
	samples2 := resample(data2, width*2)

	minVal := samples1[0]
	maxVal := samples1[0]
	for _, v := range samples1 {
		if v < minVal {
			minVal = v
		}
		if v > maxVal {
			maxVal = v
		}
	}
	for _, v := range samples2 {
		if v < minVal {
			minVal = v
		}
		if v > maxVal {
			maxVal = v
		}
	}
	if minVal > 0 {
		minVal = 0
	}
	if maxVal == minVal {
		maxVal = minVal + 1
	}

	dotRows := height * 4
	grid1 := make([][]bool, dotRows)
	grid2 := make([][]bool, dotRows)
	for i := range grid1 {
		grid1[i] = make([]bool, width*2)
		grid2[i] = make([]bool, width*2)
	}

	for x, val := range samples1 {
		y := (val - minVal) * (dotRows - 1) / (maxVal - minVal)
		if y >= dotRows {
			y = dotRows - 1
		}
		grid1[y][x] = true
	}

	for x, val := range samples2 {
		y := (val - minVal) * (dotRows - 1) / (maxVal - minVal)
		if y >= dotRows {
			y = dotRows - 1
		}
		grid2[y][x] = true
	}

	yAxisWidth := len(formatValue(maxVal)) + 1
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
			for dr := 0; dr < 4; dr++ {
				dotRow := dotR0 + dr
				if dotRow >= dotRows {
					continue
				}
				flippedRow := dotRows - 1 - dotRow
				for dc := 0; dc < 2; dc++ {
					dotCol := col*2 + dc
					if dotCol >= width*2 {
						continue
					}
					has1 := grid1[flippedRow][dotCol]
					has2 := grid2[flippedRow][dotCol]
					if has1 || has2 {
						pattern |= brailleBit(dr, dc)
					}
				}
			}

			hasAny1 := false
			hasAny2 := false
			for dr := 0; dr < 4; dr++ {
				dotRow := row*4 + dr
				if dotRow >= dotRows {
					continue
				}
				flippedRow := dotRows - 1 - dotRow
				for dc := 0; dc < 2; dc++ {
					dotCol := col*2 + dc
					if dotCol >= width*2 {
						continue
					}
					if grid1[flippedRow][dotCol] {
						hasAny1 = true
					}
					if grid2[flippedRow][dotCol] {
						hasAny2 = true
					}
				}
			}

			ch := brailleChar(pattern)
			if hasAny1 && hasAny2 {
				b.WriteString(SelectedStyle.Render(string(ch)))
			} else if hasAny1 {
				b.WriteString(CyanStyle.Render(string(ch)))
			} else if hasAny2 {
				b.WriteString(SuccessStyle.Render(string(ch)))
			} else {
				b.WriteString(" ")
			}
		}
		b.WriteString("\n")
	}

	return b.String()
}

func renderBrailleGrid(grid [][]bool, width, height int, minVal, maxVal int, styleFn func(...string) string) string {
	dotRows := height * 4
	yAxisWidth := len(formatValue(maxVal)) + 1

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
			for dr := 0; dr < 4; dr++ {
				dotRow := dotR0 + dr
				if dotRow >= dotRows {
					continue
				}
				flippedRow := dotRows - 1 - dotRow
				for dc := 0; dc < 2; dc++ {
					dotCol := col*2 + dc
					if dotCol >= width*2 {
						continue
					}
					if grid[flippedRow][dotCol] {
						pattern |= brailleBit(dr, dc)
					}
				}
			}
			if pattern == 0 {
				b.WriteString(" ")
			} else {
				b.WriteString(styleFn(string(brailleChar(pattern))))
			}
		}
		b.WriteString("\n")
	}

	return b.String()
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

func resample(data []int, targetLen int) []int {
	if len(data) <= targetLen {
		return data
	}
	result := make([]int, targetLen)
	for i := 0; i < targetLen; i++ {
		srcStart := i * len(data) / targetLen
		srcEnd := (i + 1) * len(data) / targetLen
		if srcEnd <= srcStart {
			srcEnd = srcStart + 1
		}
		if srcEnd > len(data) {
			srcEnd = len(data)
		}
		sum := 0
		for j := srcStart; j < srcEnd; j++ {
			sum += data[j]
		}
		result[i] = sum / (srcEnd - srcStart)
	}
	return result
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
