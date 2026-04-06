package screens

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/the4ofus/spacetrader-tui/internal/game"
)

type NewsScreen struct {
	gs      *game.GameState
	cursor  int
	message string
}

func NewNewsScreen(gs *game.GameState) *NewsScreen {
	return &NewsScreen{gs: gs}
}

func (s *NewsScreen) Init() tea.Cmd { return nil }

func (s *NewsScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	total := len(s.gs.NewsLog)
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, Keys.Up):
			if total > 0 {
				s.cursor = wrapCursor(s.cursor, -1, total)
			}
		case key.Matches(msg, Keys.Down):
			if total > 0 {
				s.cursor = wrapCursor(s.cursor, 1, total)
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

	if s.message != "" {
		b.WriteString("\n  " + s.message + "\n")
	}

	b.WriteString("\n" + DimStyle.Render("  j/k navigate, b bookmark, esc back"))
	return b.String()
}
