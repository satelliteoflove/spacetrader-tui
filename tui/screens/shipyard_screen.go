package screens

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/the4ofus/spacetrader-tui/internal/game"
	"github.com/the4ofus/spacetrader-tui/internal/gamedata"
	"github.com/the4ofus/spacetrader-tui/internal/shipyard"
)

type shipyardTab int

const (
	tabShips shipyardTab = iota
	tabEquipment
	tabRepair
)

type ShipyardScreen struct {
	gs         *game.GameState
	tab        shipyardTab
	cursor     int
	message    string
	ships      []gamedata.ShipDef
	equip      []gamedata.EquipDef
	confirming bool
}

func NewShipyardScreen(gs *game.GameState) *ShipyardScreen {
	return &ShipyardScreen{
		gs:    gs,
		ships: shipyard.AvailableShips(gs),
		equip: shipyard.AvailableEquipment(gs),
	}
}

func (s *ShipyardScreen) Init() tea.Cmd { return nil }

func (s *ShipyardScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if s.confirming {
		if msg, ok := msg.(tea.KeyMsg); ok {
			switch msg.String() {
			case "y":
				s.confirming = false
				if s.cursor < len(s.ships) {
					result := shipyard.BuyShip(s.gs, s.ships[s.cursor].ID)
					s.message = result.Message
					s.ships = shipyard.AvailableShips(s.gs)
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
			s.tab = tabShips
			s.cursor = 0
		case msg.String() == "2":
			s.tab = tabEquipment
			s.cursor = 0
		case msg.String() == "3":
			s.tab = tabRepair
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

func (s *ShipyardScreen) tabLen() int {
	switch s.tab {
	case tabShips:
		return len(s.ships)
	case tabEquipment:
		return len(s.equip)
	case tabRepair:
		return 4
	}
	return 0
}

func (s *ShipyardScreen) handleSelect() {
	switch s.tab {
	case tabShips:
		if s.cursor < len(s.ships) {
			ship := s.ships[s.cursor]
			s.message = SelectedStyle.Render(fmt.Sprintf("Buy %s for %d cr? (y/n)", ship.Name, ship.Price))
			s.confirming = true
			return
		}
	case tabEquipment:
		if s.cursor < len(s.equip) {
			result := shipyard.BuyEquipment(s.gs, s.equip[s.cursor].ID)
			s.message = result.Message
			s.equip = shipyard.AvailableEquipment(s.gs)
		}
	case tabRepair:
		switch s.cursor {
		case 0:
			result := shipyard.Repair(s.gs)
			s.message = result.Message
		case 1:
			result := shipyard.Refuel(s.gs)
			s.message = result.Message
		case 2:
			result := shipyard.BuyEscapePod(s.gs)
			s.message = result.Message
		case 3:
			result := shipyard.BuyInsurance(s.gs)
			s.message = result.Message
		}
	}
}

func (s *ShipyardScreen) View() string {
	var b strings.Builder

	b.WriteString(HeaderStyle.Render("  SHIPYARD  ") + "\n")
	b.WriteString(fmt.Sprintf("  Credits: %d  |  Trade-in value: %d\n\n",
		s.gs.Player.Credits, shipyard.TradeInValue(s.gs)))

	tabs := []string{"[1] Ships", "[2] Equipment", "[3] Repair/Refuel"}
	for i, t := range tabs {
		if shipyardTab(i) == s.tab {
			b.WriteString(SelectedStyle.Render(t) + "  ")
		} else {
			b.WriteString(DimStyle.Render(t) + "  ")
		}
	}
	b.WriteString("\n\n")

	switch s.tab {
	case tabShips:
		lines := make([]string, len(s.ships))
		for i, ship := range s.ships {
			lines[i] = fmt.Sprintf("%-14s %6d cr  cargo:%d  wpn:%d  shd:%d  gdt:%d",
				ship.Name, ship.Price, ship.CargoBays, ship.WeaponSlots,
				ship.ShieldSlots, ship.GadgetSlots)
		}
		RenderMenuItems(&b, lines, s.cursor)
	case tabEquipment:
		eqLines := make([]string, len(s.equip))
		for i, eq := range s.equip {
			var stat string
			switch eq.Category {
			case gamedata.EquipWeapon:
				stat = fmt.Sprintf("power:%d", eq.Power)
			case gamedata.EquipShield:
				stat = fmt.Sprintf("prot:%d", eq.Protection)
			case gamedata.EquipGadget:
				if eq.CargoBays > 0 {
					stat = fmt.Sprintf("+%d cargo", eq.CargoBays)
				} else {
					stat = fmt.Sprintf("+%s", eq.SkillBonus)
				}
			}
			eqLines[i] = fmt.Sprintf("%-20s %6d cr  %s", eq.Name, eq.Price, stat)
		}
		RenderMenuItems(&b, eqLines, s.cursor)
	case tabRepair:
		repairCost := shipyard.RepairCost(s.gs)
		refuelCost := shipyard.RefuelCost(s.gs)
		items := []string{
			fmt.Sprintf("Repair hull (%d credits)", repairCost),
			fmt.Sprintf("Refuel (%d credits)", refuelCost),
		}
		if s.gs.Player.HasEscapePod {
			items = append(items, "Escape pod [installed]")
		} else {
			items = append(items, fmt.Sprintf("Buy escape pod (%d credits)", shipyard.EscapePodPrice))
		}
		if s.gs.Player.HasInsurance {
			items = append(items, "Insurance [active]")
		} else {
			items = append(items, fmt.Sprintf("Buy insurance (%d credits)", shipyard.InsurancePrice))
		}
		RenderMenuItems(&b, items, s.cursor)
	}

	if s.message != "" {
		b.WriteString("\n  " + s.message)
	}

	b.WriteString("\n\n" + DimStyle.Render("  1/2/3 tabs, j/k navigate, enter select, esc back"))
	return b.String()
}
