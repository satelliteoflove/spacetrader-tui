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
	tabSell
	tabRepair
)

type installedItem struct {
	category gamedata.EquipCategory
	slotIdx  int
}

type ShipyardScreen struct {
	gs         *game.GameState
	tab        shipyardTab
	cursor     int
	message    string
	ships      []gamedata.ShipDef
	equip      []gamedata.EquipDef
	installed  []installedItem
	confirming bool
}

func (s *ShipyardScreen) installedEquipID(item installedItem) int {
	switch item.category {
	case gamedata.EquipWeapon:
		return s.gs.Player.Ship.Weapons[item.slotIdx]
	case gamedata.EquipShield:
		return s.gs.Player.Ship.Shields[item.slotIdx]
	case gamedata.EquipGadget:
		return s.gs.Player.Ship.Gadgets[item.slotIdx]
	}
	return 0
}

func (s *ShipyardScreen) refreshInstalled() {
	s.installed = nil
	for i := range s.gs.Player.Ship.Weapons {
		s.installed = append(s.installed, installedItem{gamedata.EquipWeapon, i})
	}
	for i := range s.gs.Player.Ship.Shields {
		s.installed = append(s.installed, installedItem{gamedata.EquipShield, i})
	}
	for i := range s.gs.Player.Ship.Gadgets {
		s.installed = append(s.installed, installedItem{gamedata.EquipGadget, i})
	}
}

func NewShipyardScreen(gs *game.GameState) *ShipyardScreen {
	s := &ShipyardScreen{
		gs:    gs,
		ships: shipyard.AvailableShips(gs),
		equip: shipyard.AvailableEquipment(gs),
	}
	s.refreshInstalled()
	return s
}

func (s *ShipyardScreen) Init() tea.Cmd { return nil }

func (s *ShipyardScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if s.confirming {
		if msg, ok := msg.(tea.KeyMsg); ok {
			switch msg.String() {
			case "y":
				s.confirming = false
				if s.tab == tabShips && s.cursor < len(s.ships) {
					result := shipyard.BuyShip(s.gs, s.ships[s.cursor].ID)
					s.message = result.Message
					s.ships = shipyard.AvailableShips(s.gs)
					s.refreshInstalled()
				} else if s.tab == tabSell && s.cursor < len(s.installed) {
					item := s.installed[s.cursor]
					result := shipyard.SellEquipment(s.gs, item.category, item.slotIdx)
					s.message = result.Message
					s.refreshInstalled()
					if s.cursor >= len(s.installed) {
						s.cursor = max(0, len(s.installed)-1)
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
			s.tab = tabShips
			s.cursor = 0
		case msg.String() == "2":
			s.tab = tabEquipment
			s.cursor = 0
		case msg.String() == "3":
			s.tab = tabSell
			s.cursor = 0
			s.refreshInstalled()
		case msg.String() == "4":
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
	case tabSell:
		return len(s.installed)
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
	case tabSell:
		if s.cursor < len(s.installed) {
			item := s.installed[s.cursor]
			equipID := s.installedEquipID(item)
			eq := s.gs.Data.Equipment[equipID]
			sellPrice := eq.Price * 3 / 4
			s.message = SelectedStyle.Render(fmt.Sprintf("Sell %s for %d cr? (y/n)", eq.Name, sellPrice))
			s.confirming = true
			return
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
	shipName := s.gs.PlayerShipDef().Name
	b.WriteString(fmt.Sprintf("  Ship: %s  |  Credits: %d  |  Trade-in: %d\n\n",
		shipName, s.gs.Player.Credits, shipyard.TradeInValue(s.gs)))

	tabs := []string{"[1] Ships", "[2] Buy", "[3] Sell", "[4] Repair/Refuel"}
	b.WriteString("  ")
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
		header := fmt.Sprintf("    %-12s %6s %5s %5s %4s %6s %6s %6s %4s",
			"SHIP", "PRICE", "CARGO", "RANGE", "HULL", "WEAPON", "SHIELD", "GADGET", "CREW")
		b.WriteString(DimStyle.Render(header) + "\n")
		b.WriteString("    " + strings.Repeat("-", 61) + "\n")
		lines := make([]string, len(s.ships))
		for i, ship := range s.ships {
			lines[i] = fmt.Sprintf("%-12s %6d %5d %5d %4d %6d %6d %6d %4d",
				ship.Name, ship.Price, ship.CargoBays, ship.Range, ship.Hull,
				ship.WeaponSlots, ship.ShieldSlots, ship.GadgetSlots, ship.CrewQuarters-1)
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
	case tabSell:
		if len(s.installed) == 0 {
			b.WriteString("  No equipment installed.\n")
		} else {
			sellLines := make([]string, len(s.installed))
			for i, item := range s.installed {
				equipID := s.installedEquipID(item)
				eq := s.gs.Data.Equipment[equipID]
				sellPrice := eq.Price * 3 / 4
				catName := "weapon"
				if item.category == gamedata.EquipShield {
					catName = "shield"
				} else if item.category == gamedata.EquipGadget {
					catName = "gadget"
				}
				sellLines[i] = fmt.Sprintf("%-20s  %-7s  sell for %d cr", eq.Name, catName, sellPrice)
			}
			RenderMenuItems(&b, sellLines, s.cursor)
		}
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

	b.WriteString("\n\n" + DimStyle.Render("  1/2/3/4 tabs, j/k navigate, enter select, esc back"))
	return b.String()
}
