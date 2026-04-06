package screens

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/the4ofus/spacetrader-tui/internal/game"
)

type NewsScreen struct {
	gs       *game.GameState
	cursor   int
	message  string
	briefing *game.NewsBriefing
}

func NewNewsScreen(gs *game.GameState) *NewsScreen {
	return &NewsScreen{gs: gs}
}

func (s *NewsScreen) Init() tea.Cmd { return nil }

func (s *NewsScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	total := len(s.gs.NewsLog)

	if s.briefing != nil {
		if msg, ok := msg.(tea.KeyMsg); ok {
			switch {
			case msg.String() == "m":
				entry := s.gs.NewsLog[s.reverseIdx()]
				s.briefing = nil
				return s, func() tea.Msg {
					return NavigateMsg{Screen: ScreenGalacticChart, SelectedSystem: entry.SystemIdx}
				}
			case msg.String() == "t":
				s.briefing = nil
				return s, func() tea.Msg {
					return NavigateMsg{Screen: ScreenMarket}
				}
			case msg.String() == "b":
				entry := s.gs.NewsLog[s.reverseIdx()]
				added := s.gs.ToggleBookmark(entry.SystemIdx, entry.Headline)
				if added {
					s.message = SuccessStyle.Render(fmt.Sprintf("Bookmarked %s", entry.System))
				} else {
					s.message = DimStyle.Render(fmt.Sprintf("Removed bookmark for %s", entry.System))
				}
			default:
				s.briefing = nil
				s.message = ""
			}
		}
		return s, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, Keys.Up):
			if total > 0 {
				s.cursor = wrapCursor(s.cursor, -1, total)
				s.message = ""
			}
		case key.Matches(msg, Keys.Down):
			if total > 0 {
				s.cursor = wrapCursor(s.cursor, 1, total)
				s.message = ""
			}
		case key.Matches(msg, Keys.Enter):
			if total > 0 {
				entry := s.gs.NewsLog[s.reverseIdx()]
				brief := game.GenerateNewsBriefing(s.gs, entry)
				s.briefing = &brief
			}
		case msg.String() == "b":
			if total > 0 {
				entry := s.gs.NewsLog[s.reverseIdx()]
				added := s.gs.ToggleBookmark(entry.SystemIdx, entry.Headline)
				if added {
					s.message = SuccessStyle.Render(fmt.Sprintf("Bookmarked %s", entry.System))
				} else {
					s.message = DimStyle.Render(fmt.Sprintf("Removed bookmark for %s", entry.System))
				}
			}
		case key.Matches(msg, Keys.Back):
			return s, func() tea.Msg { return NavigateMsg{Screen: ScreenSystem} }
		}
	}
	return s, nil
}

func (s *NewsScreen) reverseIdx() int {
	return len(s.gs.NewsLog) - 1 - s.cursor
}

func (s *NewsScreen) View() string {
	var b strings.Builder

	b.WriteString(HeaderStyle.Render("  RECENT NEWS  ") + "\n\n")

	total := len(s.gs.NewsLog)
	if total == 0 {
		b.WriteString("  No news reports yet.\n")
		b.WriteString("\n" + DimStyle.Render("  esc back"))
		return b.String()
	}

	if s.cursor >= total {
		s.cursor = total - 1
	}

	pageSize := 14
	start := s.cursor - pageSize/2
	if start < 0 {
		start = 0
	}
	end := start + pageSize
	if end > total {
		end = total
		start = end - pageSize
		if start < 0 {
			start = 0
		}
	}

	for i := start; i < end; i++ {
		revIdx := total - 1 - i
		entry := s.gs.NewsLog[revIdx]
		age := s.gs.Day - entry.Day
		var ageStr string
		switch {
		case age == 0:
			ageStr = SuccessStyle.Render("today")
		case age == 1:
			ageStr = "1 day ago"
		case age <= 5:
			ageStr = fmt.Sprintf("%d days ago", age)
		default:
			ageStr = DimStyle.Render(fmt.Sprintf("%d days ago", age))
		}

		bookmarkMarker := " "
		if s.gs.IsBookmarked(entry.SystemIdx) {
			bookmarkMarker = SelectedStyle.Render("!")
		}

		line := fmt.Sprintf("%s %s %s", bookmarkMarker, DimStyle.Render(fmt.Sprintf("[%-10s]", ageStr)), entry.Headline)
		if i == s.cursor {
			b.WriteString(SelectedStyle.Render(">") + line + "\n")
		} else {
			b.WriteString(" " + line + "\n")
		}
	}

	if s.briefing != nil {
		b.WriteString("\n")
		b.WriteString(s.renderBriefing())
	}

	if s.message != "" {
		b.WriteString("\n  " + s.message + "\n")
	}

	if s.briefing != nil {
		b.WriteString("\n" + DimStyle.Render("  t market, m galactic map, b bookmark, any key back"))
	} else {
		b.WriteString("\n" + DimStyle.Render("  j/k navigate, enter details, b bookmark, esc back"))
	}
	return b.String()
}

func (s *NewsScreen) renderBriefing() string {
	var b strings.Builder
	brief := s.briefing

	b.WriteString(CyanStyle.Render(fmt.Sprintf("  -- %s MARKET REPORT --", strings.ToUpper(brief.SystemName))) + "\n\n")

	wrapped := WordWrap(brief.Blurb, 72)
	for _, line := range strings.Split(wrapped, "\n") {
		b.WriteString("  " + line + "\n")
	}

	if brief.EventActive && len(brief.PriceLines) > 0 {
		b.WriteString("\n")
		for _, line := range brief.PriceLines {
			b.WriteString("  " + line + "\n")
		}
	}

	for _, alert := range brief.CargoAlerts {
		b.WriteString("\n  " + SelectedStyle.Render("** "+alert))
	}

	if brief.SecurityWarning != "" {
		b.WriteString("\n  " + DangerStyle.Render(brief.SecurityWarning))
	}

	b.WriteString("\n")
	rangeStr := DangerStyle.Render("out of range")
	if brief.InRange {
		rangeStr = SuccessStyle.Render("in range")
	}
	b.WriteString(fmt.Sprintf("\n  Distance: %.1f parsecs (%s)\n", brief.Distance, rangeStr))

	if brief.EventActive {
		b.WriteString("  " + DimStyle.Render("Prices are estimates and may vary.") + "\n")
	}

	return b.String()
}
