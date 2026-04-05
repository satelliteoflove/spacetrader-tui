package screens

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/the4ofus/spacetrader-tui/internal/game"
	"github.com/the4ofus/spacetrader-tui/internal/gamedata"
)

type SystemScreen struct {
	gs        *game.GameState
	cursor    int
	items     []menuItem
	headlines []string
}

type menuItem struct {
	label  string
	screen ScreenType
}

func NewSystemScreen(gs *game.GameState) *SystemScreen {
	return NewSystemScreenWithCursor(gs, 0)
}

func NewSystemScreenWithCursor(gs *game.GameState, cursor int) *SystemScreen {
	items := []menuItem{
		{"Market", ScreenMarket},
		{"Short-Range Chart", ScreenChart},
		{"Shipyard", ScreenShipyard},
		{"Bank", ScreenBank},
		{"Personnel", ScreenPersonnel},
		{"Galactic Chart", ScreenGalacticChart},
		{"Status", ScreenStatus},
	}

	if gs.Player.Credits >= 500000 && gs.Player.LoanBalance == 0 {
		items = append(items, menuItem{"Buy Moon and Retire!", ScreenGameOver})
	}

	if cursor >= len(items) {
		cursor = 0
	}

	headlines := game.GenerateNewspaper(gs)
	return &SystemScreen{gs: gs, cursor: cursor, items: items, headlines: headlines}
}

func (s *SystemScreen) Init() tea.Cmd { return nil }

func (s *SystemScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, Keys.Up):
			s.cursor = wrapCursor(s.cursor, -1, len(s.items))
		case key.Matches(msg, Keys.Down):
			s.cursor = wrapCursor(s.cursor, 1, len(s.items))
		case key.Matches(msg, Keys.Enter):
			target := s.items[s.cursor].screen
			if target == ScreenGameOver {
				s.gs.Player.Credits -= 500000
				s.gs.Player.MoonPurchased = true
				s.gs.EndStatus = game.StatusRetired
			}
			cursor := s.cursor
			return s, func() tea.Msg { return NavigateMsg{Screen: target, RestoreCursor: cursor} }
		case msg.String() == "s":
			cursor := s.cursor
			return s, func() tea.Msg { return NavigateMsg{Screen: ScreenSave, RestoreCursor: cursor} }
		}
	}
	return s, nil
}

func (s *SystemScreen) View() string {
	var b strings.Builder

	sys := s.gs.CurrentSystem()
	shipDef := s.gs.PlayerShipDef()

	b.WriteString(HeaderStyle.Render(fmt.Sprintf("  %s  ", sys.Name)) + "\n")
	b.WriteString(fmt.Sprintf("  Tech: %s  |  Gov: %s  |  Resource: %s\n",
		sys.TechLevel, sys.PoliticalSystem, sys.Resource))
	b.WriteString(fmt.Sprintf("  Credits: %d  |  Ship: %s  |  Hull: %d/%d  |  Fuel: %d/%d  |  Day: %d\n",
		s.gs.Player.Credits, shipDef.Name, s.gs.Player.Ship.Hull, shipDef.Hull,
		s.gs.Player.Ship.Fuel, shipDef.Range, s.gs.Day))

	if s.gs.Player.LoanBalance > 0 {
		b.WriteString(DangerStyle.Render(fmt.Sprintf("  Debt: %d", s.gs.Player.LoanBalance)) + "\n")
	}

	record := gamedata.PoliceRecordToTier(s.gs.Player.PoliceRecord)
	rep := gamedata.ReputationToTier(s.gs.Player.Reputation)
	b.WriteString(fmt.Sprintf("  Record: %s  |  Reputation: %s\n", record, rep))

	b.WriteString("\n")

	for i, item := range s.items {
		if i == s.cursor {
			b.WriteString(fmt.Sprintf("  %s\n", SelectedStyle.Render("> "+item.label)))
		} else {
			b.WriteString(fmt.Sprintf("    %s\n", NormalStyle.Render(item.label)))
		}
	}

	if len(s.headlines) > 0 {
		b.WriteString("\n" + DimStyle.Render("  --- News ---") + "\n")
		for _, h := range s.headlines {
			b.WriteString("  " + DimStyle.Render(h) + "\n")
		}
	}

	b.WriteString("\n" + DimStyle.Render("  j/k navigate, enter select, s save, ctrl+c quit"))
	return b.String()
}
