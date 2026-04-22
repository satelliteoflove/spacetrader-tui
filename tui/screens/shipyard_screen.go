package screens

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/the4ofus/spacetrader-tui/internal/formula"
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

type confirmPhase int

const (
	phaseNone confirmPhase = iota
	phaseSellEquip
	phaseShipPreview
	phaseCrewPick
)

type installedItem struct {
	category gamedata.EquipCategory
	slotIdx  int
}

type crewOption struct {
	crewIdx int
	merc    game.Mercenary
}

type extraService struct {
	line string
	buy  func(*game.GameState) shipyard.Result
}

type ShipyardScreen struct {
	gs            *game.GameState
	tab           shipyardTab
	cursor        int
	message       string
	ships         []gamedata.ShipDef
	equip         []gamedata.EquipDef
	installed     []installedItem
	phase         confirmPhase
	preview       *shipyard.ShipPurchasePreview
	crewOptions   []crewOption
	crewCursor    int
	crewDismissed map[int]bool
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
	switch s.phase {
	case phaseSellEquip:
		return s.updateSellConfirm(msg)
	case phaseShipPreview:
		return s.updateShipPreview(msg)
	case phaseCrewPick:
		return s.updateCrewPick(msg)
	}
	return s.updateNormal(msg)
}

func (s *ShipyardScreen) updateNormal(msg tea.Msg) (tea.Model, tea.Cmd) {
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

func (s *ShipyardScreen) updateSellConfirm(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(tea.KeyMsg); ok {
		switch msg.String() {
		case "y":
			s.phase = phaseNone
			if s.cursor < len(s.installed) {
				item := s.installed[s.cursor]
				result := shipyard.SellEquipment(s.gs, item.category, item.slotIdx)
				s.message = result.Message
				s.refreshInstalled()
				if s.cursor >= len(s.installed) {
					s.cursor = max(0, len(s.installed)-1)
				}
			}
		default:
			s.phase = phaseNone
			s.message = ""
		}
	}
	return s, nil
}

func (s *ShipyardScreen) updateShipPreview(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(tea.KeyMsg); ok {
		switch msg.String() {
		case "y":
			if s.preview.CrewMustCut > 0 {
				s.enterCrewPick()
				return s, nil
			}
			s.executePurchase(nil)
		case "n", "esc", "q":
			s.phase = phaseNone
			s.preview = nil
			s.message = ""
		}
	}
	return s, nil
}

func (s *ShipyardScreen) updateCrewPick(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(tea.KeyMsg); ok {
		switch {
		case key.Matches(msg, Keys.Up):
			s.crewCursor = wrapCursor(s.crewCursor, -1, len(s.crewOptions))
		case key.Matches(msg, Keys.Down):
			s.crewCursor = wrapCursor(s.crewCursor, 1, len(s.crewOptions))
		case msg.String() == " ":
			idx := s.crewOptions[s.crewCursor].crewIdx
			if s.crewDismissed[idx] {
				delete(s.crewDismissed, idx)
			} else if len(s.crewDismissed) < s.preview.CrewMustCut {
				s.crewDismissed[idx] = true
			}
		case msg.String() == "y":
			if len(s.crewDismissed) == s.preview.CrewMustCut {
				var dismissed []int
				for idx := range s.crewDismissed {
					dismissed = append(dismissed, idx)
				}
				s.executePurchase(dismissed)
			}
		case msg.String() == "n" || msg.String() == "esc" || msg.String() == "q":
			s.phase = phaseNone
			s.preview = nil
			s.crewDismissed = nil
			s.message = ""
		}
	}
	return s, nil
}

func (s *ShipyardScreen) enterCrewPick() {
	s.phase = phaseCrewPick
	s.crewCursor = 0
	s.crewDismissed = make(map[int]bool)
	s.crewOptions = nil
	for i, m := range s.gs.Player.Crew {
		if !m.IsQuest {
			s.crewOptions = append(s.crewOptions, crewOption{crewIdx: i, merc: m})
		}
	}
}

func (s *ShipyardScreen) executePurchase(dismissCrew []int) {
	result := shipyard.BuyShip(s.gs, s.preview.NewShip.ID, dismissCrew)
	s.message = result.Message
	s.phase = phaseNone
	s.preview = nil
	s.crewDismissed = nil
	s.ships = shipyard.AvailableShips(s.gs)
	s.refreshInstalled()
}

func (s *ShipyardScreen) tabLen() int {
	switch s.tab {
	case tabShips:
		return len(s.ships)
	case tabEquipment:
		return len(s.equip) + len(s.extraServices())
	case tabSell:
		return len(s.installed)
	case tabRepair:
		return 2
	}
	return 0
}

func (s *ShipyardScreen) extraServices() []extraService {
	var podLine string
	if s.gs.Player.HasEscapePod {
		podLine = fmt.Sprintf("%-32s[installed]", "Escape pod")
	} else {
		podLine = fmt.Sprintf("%-20s %6d cr  survive if ship destroyed", "Escape pod", shipyard.EscapePodPrice)
	}

	insuranceDesc := "recover 75% of ship value"
	if !s.gs.Player.HasEscapePod {
		insuranceDesc = "recover 75%; needs pod"
	}
	var insuranceLine string
	if s.gs.Player.HasInsurance {
		insuranceLine = fmt.Sprintf("%-32s[active]", "Insurance")
	} else {
		premium := game.InsuranceBasePremium(s.gs)
		insuranceLine = fmt.Sprintf("%-20s %5d cr/d  %s", "Insurance", premium, insuranceDesc)
	}

	return []extraService{
		{line: podLine, buy: shipyard.BuyEscapePod},
		{line: insuranceLine, buy: shipyard.BuyInsurance},
	}
}

func (s *ShipyardScreen) handleSelect() {
	switch s.tab {
	case tabShips:
		if s.cursor < len(s.ships) {
			ship := s.ships[s.cursor]
			preview := shipyard.PreviewShipPurchase(s.gs, ship.ID)
			if preview.Error != "" {
				s.message = preview.Error
				return
			}
			s.preview = &preview
			s.phase = phaseShipPreview
			s.message = ""
		}
	case tabEquipment:
		if s.cursor < len(s.equip) {
			result := shipyard.BuyEquipment(s.gs, s.equip[s.cursor].ID)
			s.message = result.Message
			s.equip = shipyard.AvailableEquipment(s.gs)
		} else {
			services := s.extraServices()
			idx := s.cursor - len(s.equip)
			if idx >= 0 && idx < len(services) {
				result := services[idx].buy(s.gs)
				s.message = result.Message
			}
		}
	case tabSell:
		if s.cursor < len(s.installed) {
			item := s.installed[s.cursor]
			equipID := s.installedEquipID(item)
			eq := s.gs.Data.Equipment[equipID]
			sellPrice := eq.Price * 3 / 4
			prompt := SelectedStyle.Render(fmt.Sprintf("Sell %s for %d cr? (y/n)", eq.Name, sellPrice))
			if item.category == gamedata.EquipWeapon && len(s.gs.Player.Ship.Weapons) == 1 {
				prompt = DangerStyle.Render("Warning: this is your last weapon.") + "\n  " + prompt
			} else if item.category == gamedata.EquipShield && len(s.gs.Player.Ship.Shields) == 1 {
				prompt = DangerStyle.Render("Warning: this is your last shield.") + "\n  " + prompt
			}
			s.message = prompt
			s.phase = phaseSellEquip
		}
	case tabRepair:
		switch s.cursor {
		case 0:
			result := shipyard.Repair(s.gs)
			s.message = result.Message
		case 1:
			result := shipyard.Refuel(s.gs)
			s.message = result.Message
		}
	}
}

func (s *ShipyardScreen) View() string {
	var b strings.Builder

	b.WriteString(HeaderStyle.Render("  SHIPYARD  ") + "\n")
	shipName := s.gs.PlayerShipDef().Name
	b.WriteString(fmt.Sprintf("  Ship: %s  |  Credits: %d  |  Trade-in: %d\n\n",
		shipName, s.gs.Player.Credits, shipyard.ShipHullTradeIn(s.gs)))

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

	switch s.phase {
	case phaseShipPreview:
		s.renderPreview(&b)
	case phaseCrewPick:
		s.renderCrewPick(&b)
	default:
		s.renderTab(&b)
	}

	if s.phase == phaseNone || s.phase == phaseSellEquip {
		if s.message != "" {
			b.WriteString("\n  " + s.message)
		}
		b.WriteString("\n\n" + DimStyle.Render("  1/2/3/4 tabs, j/k navigate, enter select, esc back, ? help"))
	}

	return b.String()
}

func (s *ShipyardScreen) renderTab(b *strings.Builder) {
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
		RenderMenuItems(b, lines, s.cursor)
	case tabEquipment:
		services := s.extraServices()
		eqLines := make([]string, 0, len(s.equip)+len(services))
		for _, eq := range s.equip {
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
			eqLines = append(eqLines, fmt.Sprintf("%-20s %6d cr  %s", eq.Name, eq.Price, stat))
		}
		for _, svc := range services {
			eqLines = append(eqLines, svc.line)
		}
		RenderMenuItems(b, eqLines, s.cursor)
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
			RenderMenuItems(b, sellLines, s.cursor)
		}
	case tabRepair:
		repairCost := shipyard.RepairCost(s.gs)
		refuelCost := shipyard.RefuelCost(s.gs)
		items := []string{
			fmt.Sprintf("Repair hull (%d credits)", repairCost),
			fmt.Sprintf("Refuel (%d credits)", refuelCost),
		}
		RenderMenuItems(b, items, s.cursor)
	}
}

func (s *ShipyardScreen) renderPreview(b *strings.Builder) {
	p := s.preview
	b.WriteString(SelectedStyle.Render(fmt.Sprintf("  === Buy %s ===", p.NewShip.Name)) + "\n\n")

	hasTransferInfo := false

	if len(p.Weapons.Kept) > 0 || len(p.Weapons.Sold) > 0 ||
		len(p.Shields.Kept) > 0 || len(p.Shields.Sold) > 0 ||
		len(p.Gadgets.Kept) > 0 || len(p.Gadgets.Sold) > 0 {
		b.WriteString("  Equipment:\n")
		hasTransferInfo = true
		s.renderEquipCategory(b, "Weapons", p.Weapons)
		s.renderEquipCategory(b, "Shields", p.Shields)
		s.renderEquipCategory(b, "Gadgets", p.Gadgets)
	}

	crewCount := len(s.gs.Player.Crew)
	if crewCount > 0 {
		hasTransferInfo = true
		if p.CrewMustCut > 0 {
			b.WriteString(fmt.Sprintf("\n  Crew: must dismiss %d of %d members\n", p.CrewMustCut, crewCount))
		} else {
			b.WriteString("\n  Crew: all members transfer\n")
		}
	}

	if hasTransferInfo {
		b.WriteString("\n")
	}

	b.WriteString(DimStyle.Render("  ----------------------------") + "\n")
	b.WriteString(fmt.Sprintf("  Ship price:     %7d cr\n", p.NewShip.Price))
	b.WriteString(fmt.Sprintf("  Hull trade-in: -%7d cr\n", p.HullTradeIn))

	equipSold := p.Weapons.SoldValue + p.Shields.SoldValue + p.Gadgets.SoldValue
	if equipSold > 0 {
		b.WriteString(fmt.Sprintf("  Equip sold:    -%7d cr\n", equipSold))
	}

	b.WriteString(DimStyle.Render("  ----------------------------") + "\n")
	if p.NetCost >= 0 {
		b.WriteString(SelectedStyle.Render(fmt.Sprintf("  You pay:        %7d cr", p.NetCost)) + "\n")
	} else {
		b.WriteString(SuccessStyle.Render(fmt.Sprintf("  You receive:    %7d cr", -p.NetCost)) + "\n")
	}
	b.WriteString(fmt.Sprintf("  Credits after:  %7d cr\n", s.gs.Player.Credits-p.NetCost))

	if p.CrewMustCut > 0 {
		b.WriteString("\n" + SelectedStyle.Render("  Proceed to crew selection? (y/n)"))
	} else {
		b.WriteString("\n" + SelectedStyle.Render("  Confirm purchase? (y/n)"))
	}
}

func (s *ShipyardScreen) renderEquipCategory(b *strings.Builder, label string, eq shipyard.EquipSummary) {
	if len(eq.Kept) == 0 && len(eq.Sold) == 0 {
		return
	}
	b.WriteString(fmt.Sprintf("    %s: ", label))
	var parts []string
	for _, eid := range eq.Kept {
		parts = append(parts, SuccessStyle.Render(s.gs.Data.Equipment[eid].Name+" (kept)"))
	}
	for _, eid := range eq.Sold {
		e := s.gs.Data.Equipment[eid]
		parts = append(parts, DimStyle.Render(fmt.Sprintf("%s (sold: %d cr)", e.Name, e.Price*3/4)))
	}
	b.WriteString(strings.Join(parts, ", ") + "\n")
}

func (s *ShipyardScreen) renderCrewPick(b *strings.Builder) {
	p := s.preview
	b.WriteString(SelectedStyle.Render(fmt.Sprintf("  === Dismiss %d crew member(s) ===", p.CrewMustCut)) + "\n\n")

	for i, opt := range s.crewOptions {
		check := "[ ]"
		if s.crewDismissed[opt.crewIdx] {
			check = "[x]"
		}
		skills := fmt.Sprintf("P:%d F:%d T:%d E:%d",
			opt.merc.Skills[formula.SkillPilot],
			opt.merc.Skills[formula.SkillFighter],
			opt.merc.Skills[formula.SkillTrader],
			opt.merc.Skills[formula.SkillEngineer])
		line := fmt.Sprintf("%s %-12s  %s", check, opt.merc.Name, skills)
		if i == s.crewCursor {
			b.WriteString("  " + SelectedStyle.Render("> "+line) + "\n")
		} else {
			b.WriteString("    " + NormalStyle.Render(line) + "\n")
		}
	}

	selected := len(s.crewDismissed)
	b.WriteString(fmt.Sprintf("\n  Selected: %d / %d", selected, p.CrewMustCut))

	if selected == p.CrewMustCut {
		b.WriteString("\n" + SelectedStyle.Render("  Press y to confirm, esc to cancel"))
	} else {
		b.WriteString("\n" + DimStyle.Render("  space toggle, j/k navigate, esc cancel"))
	}
}

func (s *ShipyardScreen) HelpTitle() string { return "Shipyard" }

func (s *ShipyardScreen) HelpGroups() []KeyGroup {
	return []KeyGroup{
		{
			Title: "Tabs",
			Bindings: []KeyBinding{
				{Keys: "1", Desc: "Ships (buy new ship)"},
				{Keys: "2", Desc: "Buy equipment"},
				{Keys: "3", Desc: "Sell equipment"},
				{Keys: "4", Desc: "Repair / refuel"},
			},
		},
		{
			Title: "Navigation",
			Bindings: []KeyBinding{
				{Keys: "j/k or arrows", Desc: "Move cursor"},
				{Keys: "enter", Desc: "Select / confirm"},
				{Keys: "y / n", Desc: "Confirm prompts"},
			},
		},
	}
}
