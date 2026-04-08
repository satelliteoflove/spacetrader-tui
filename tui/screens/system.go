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
		{"Navigation", ScreenGalacticList},
		{"Galaxy News", ScreenNews},
		{"Status", ScreenStatus},
		{"Shipyard", ScreenShipyard},
		{"Personnel", ScreenPersonnel},
		{"Bank", ScreenBank},
		{"Trader's Guide", ScreenGuide},
		{"Settings", ScreenSettings},
	}

	if gs.Player.Credits >= 500000 && gs.Player.LoanBalance == 0 && gs.QuestState(game.QuestMoonForSale) == game.QuestAvailable {
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
			if target == ScreenGameOver && !s.gs.Player.MoonPurchased {
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

	resLabel := shortResource(sys.Resource)
	if resLabel == "" {
		resLabel = "None"
	}
	b.WriteString(fmt.Sprintf("  Tech: %s  |  Gov: %s  |  Size: %s\n",
		sys.TechLevel, sys.PoliticalSystem, sys.Size))
	b.WriteString(fmt.Sprintf("  Specialty: %s\n", resLabel))
	b.WriteString(DimStyle.Render("  " + strings.Repeat("-", 40)) + "\n")

	record := gamedata.PoliceRecordToTier(s.gs.Player.PoliceRecord)
	rep := gamedata.ReputationToTier(s.gs.Player.Reputation)
	b.WriteString(fmt.Sprintf("  Ship: %s  |  Record: %s  |  Rep: %s\n",
		shipDef.Name, record, rep))

	if s.gs.Player.LoanBalance > 0 {
		b.WriteString(DangerStyle.Render(fmt.Sprintf("  Debt: %d cr (10%% interest per warp)", s.gs.Player.LoanBalance)) + "\n")
	}

	if len(s.headlines) > 0 {
		masthead := game.SystemMasthead(s.gs)
		b.WriteString("\n" + CyanStyle.Render(fmt.Sprintf("  --- %s ---", masthead)) + "\n")
		for _, h := range s.headlines {
			b.WriteString("  " + h + "\n")
		}
	}

	b.WriteString("\n")

	labels := make([]string, len(s.items))
	for i, item := range s.items {
		labels[i] = item.label
	}
	RenderMenuItems(&b, labels, s.cursor)

	b.WriteString("\n" + DimStyle.Render("  j/k navigate, enter select, s save, ctrl+c quit"))
	return b.String()
}
