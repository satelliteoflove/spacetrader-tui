package screens

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/the4ofus/spacetrader-tui/internal/game"
)

type personnelTab int

const (
	tabCrew personnelTab = iota
	tabHire
)

type PersonnelScreen struct {
	gs         *game.GameState
	tab        personnelTab
	cursor     int
	message    string
	available  []int
	confirming bool
}

func NewPersonnelScreen(gs *game.GameState) *PersonnelScreen {
	return &PersonnelScreen{
		gs:        gs,
		available: game.AvailableMercenaries(gs),
	}
}

func (s *PersonnelScreen) Init() tea.Cmd { return nil }

func (s *PersonnelScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if s.confirming {
		if msg, ok := msg.(tea.KeyMsg); ok {
			switch msg.String() {
			case "y":
				s.confirming = false
				if s.cursor < len(s.gs.Player.Crew) {
					ok, msg := game.FireMercenary(s.gs, s.cursor)
					s.message = msg
					if ok && s.cursor >= len(s.gs.Player.Crew) {
						s.cursor = max(0, len(s.gs.Player.Crew)-1)
					}
				}
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
		case msg.String() == "1":
			s.tab = tabCrew
			s.cursor = 0
		case msg.String() == "2":
			s.tab = tabHire
			s.cursor = 0
		case key.Matches(msg, Keys.Up):
			s.cursor = wrapCursor(s.cursor, -1, s.tabLen())
		case key.Matches(msg, Keys.Down):
			s.cursor = wrapCursor(s.cursor, 1, s.tabLen())
		case key.Matches(msg, Keys.Enter):
			s.handleSelect()
		case key.Matches(msg, Keys.Back):
			return s, func() tea.Msg { return NavigateMsg{Screen: ScreenSystem} }
		}
	}
	return s, nil
}

func (s *PersonnelScreen) tabLen() int {
	switch s.tab {
	case tabCrew:
		return len(s.gs.Player.Crew)
	case tabHire:
		return len(s.available)
	}
	return 0
}

func (s *PersonnelScreen) handleSelect() {
	switch s.tab {
	case tabCrew:
		if s.cursor < len(s.gs.Player.Crew) {
			m := s.gs.Player.Crew[s.cursor]
			if m.IsQuest {
				s.message = DimStyle.Render(fmt.Sprintf("%s is a passenger and cannot be dismissed.", m.Name))
				return
			}
			s.message = SelectedStyle.Render(fmt.Sprintf("Fire %s? (y/n)", m.Name))
			s.confirming = true
			return
		}
	case tabHire:
		if s.cursor < len(s.available) {
			mercIdx := s.available[s.cursor]
			ok, msg := game.HireMercenary(s.gs, mercIdx)
			s.message = msg
			if ok {
				s.available = game.AvailableMercenaries(s.gs)
				if s.cursor >= len(s.available) {
					s.cursor = max(0, len(s.available)-1)
				}
			}
		}
	}
}

func (s *PersonnelScreen) View() string {
	var b strings.Builder

	shipDef := s.gs.Data.Ships[s.gs.Player.Ship.TypeID]
	maxCrew := shipDef.CrewQuarters - 1

	b.WriteString(HeaderStyle.Render("  PERSONNEL  ") + "\n")
	b.WriteString(fmt.Sprintf("  Crew: %d/%d  |  Credits: %d\n\n",
		len(s.gs.Player.Crew), maxCrew, s.gs.Player.Credits))

	tabs := []string{"[1] Current Crew", "[2] Hire"}
	b.WriteString("  ")
	for i, t := range tabs {
		if personnelTab(i) == s.tab {
			b.WriteString(SelectedStyle.Render(t) + "  ")
		} else {
			b.WriteString(DimStyle.Render(t) + "  ")
		}
	}
	b.WriteString("\n\n")

	skillHeader := fmt.Sprintf("  %-12s %4s %4s %4s %4s %8s", "NAME", "PLT", "FGT", "TRD", "ENG", "WAGE")
	b.WriteString(DimStyle.Render(skillHeader) + "\n")
	b.WriteString("  " + strings.Repeat("-", 44) + "\n")

	switch s.tab {
	case tabCrew:
		if len(s.gs.Player.Crew) == 0 {
			b.WriteString("  No crew members.\n")
		}
		crewLines := make([]string, len(s.gs.Player.Crew))
		for i, m := range s.gs.Player.Crew {
			wageStr := fmt.Sprintf("%d/d", m.Wage())
			if m.IsQuest {
				wageStr = "free"
			}
			tag := ""
			if m.IsQuest {
				tag = " *"
			}
			crewLines[i] = fmt.Sprintf("%-12s %4d %4d %4d %4d %8s%s",
				m.Name, m.Skills[0], m.Skills[1], m.Skills[2], m.Skills[3], wageStr, tag)
		}
		RenderMenuItems(&b, crewLines, s.cursor)
		if len(s.gs.Player.Crew) > 0 {
			hasRegular := false
			for _, m := range s.gs.Player.Crew {
				if !m.IsQuest {
					hasRegular = true
					break
				}
			}
			if hasRegular {
				b.WriteString("\n" + DimStyle.Render("  enter to fire (* = passenger, cannot dismiss)"))
			} else {
				b.WriteString("\n" + DimStyle.Render("  * = passenger, cannot dismiss"))
			}
		}
	case tabHire:
		if len(s.available) == 0 {
			b.WriteString("  No mercenaries available here.\n")
		}
		hireLines := make([]string, len(s.available))
		for i, mercIdx := range s.available {
			m := s.gs.Mercenaries[mercIdx]
			hireLines[i] = fmt.Sprintf("%-12s %4d %4d %4d %4d %6d/d",
				m.Name, m.Skills[0], m.Skills[1], m.Skills[2], m.Skills[3], m.Wage())
		}
		RenderMenuItems(&b, hireLines, s.cursor)
		if len(s.available) > 0 {
			b.WriteString("\n" + DimStyle.Render("  enter to hire"))
		}
	}

	if s.message != "" {
		b.WriteString("\n  " + s.message)
	}

	b.WriteString("\n\n" + DimStyle.Render("  1/2 tabs, j/k navigate, enter select, esc back"))
	return b.String()
}
