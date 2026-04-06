package screens

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/the4ofus/spacetrader-tui/internal/game"
	"github.com/the4ofus/spacetrader-tui/internal/gamedata"
)

type StatusScreen struct {
	gs *game.GameState
}

func NewStatusScreen(gs *game.GameState) *StatusScreen {
	return &StatusScreen{gs: gs}
}

func (s *StatusScreen) Init() tea.Cmd { return nil }

func (s *StatusScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if key.Matches(msg, Keys.Back) {
			return s, func() tea.Msg { return NavigateMsg{Screen: ScreenSystem} }
		}
	}
	return s, nil
}

func (s *StatusScreen) View() string {
	var b strings.Builder
	p := &s.gs.Player
	shipDef := s.gs.PlayerShipDef()

	b.WriteString(HeaderStyle.Render("  COMMANDER STATUS  ") + "\n\n")

	div := DimStyle.Render("  " + strings.Repeat("-", 40)) + "\n"

	b.WriteString(fmt.Sprintf("  Name: %s\n", p.Name))
	b.WriteString(fmt.Sprintf("  Difficulty: %s\n", s.gs.Difficulty))
	b.WriteString(fmt.Sprintf("  Day: %d\n", s.gs.Day))
	b.WriteString(fmt.Sprintf("  Credits: %d\n", p.Credits))
	if p.LoanBalance > 0 {
		b.WriteString(DangerStyle.Render(fmt.Sprintf("  Debt: %d", p.LoanBalance)) + "\n")
	}

	skillNames := []string{"Pilot", "Fighter", "Trader", "Engineer"}
	b.WriteString("  Skills:\n")
	for i, name := range skillNames {
		b.WriteString(fmt.Sprintf("    %-10s %d\n", name, p.Skills[i]))
	}

	record := gamedata.PoliceRecordToTier(p.PoliceRecord)
	rep := gamedata.ReputationToTier(p.Reputation)
	b.WriteString(fmt.Sprintf("  Record: %s (%d)  |  Rep: %s (%d)\n", record, p.PoliceRecord, rep, p.Reputation))

	b.WriteString("\n" + div + "\n")

	b.WriteString(fmt.Sprintf("  Ship: %s\n", shipDef.Name))
	b.WriteString(fmt.Sprintf("  Hull: %d/%d  |  Fuel: %d/%d  |  Cargo: %d/%d\n",
		p.Ship.Hull, shipDef.Hull, p.Ship.Fuel, shipDef.Range, p.TotalCargo(), shipDef.CargoBays))

	if len(p.Ship.Weapons) > 0 {
		b.WriteString("  Weapons:")
		for _, w := range p.Ship.Weapons {
			b.WriteString(" " + s.gs.Data.Equipment[w].Name)
		}
		b.WriteString("\n")
	}
	if len(p.Ship.Shields) > 0 {
		b.WriteString("  Shields:")
		for _, sh := range p.Ship.Shields {
			b.WriteString(" " + s.gs.Data.Equipment[sh].Name)
		}
		b.WriteString("\n")
	}
	if len(p.Ship.Gadgets) > 0 {
		b.WriteString("  Gadgets:")
		for _, g := range p.Ship.Gadgets {
			b.WriteString(" " + s.gs.Data.Equipment[g].Name)
		}
		b.WriteString("\n")
	}

	if p.HasEscapePod {
		b.WriteString("  Escape pod: installed")
		if p.HasInsurance {
			b.WriteString("  |  Insurance: active")
		}
		b.WriteString("\n")
	}

	if len(p.Crew) > 0 {
		b.WriteString("  Crew:")
		for _, m := range p.Crew {
			b.WriteString(fmt.Sprintf(" %s (%d/d)", m.Name, m.Wage))
		}
		b.WriteString("\n")
	}

	b.WriteString("\n" + div + "\n")

	if p.TotalCargo() > 0 {
		b.WriteString(CyanStyle.Render("  Cargo Hold") + "\n")
		for i := 0; i < game.NumGoods; i++ {
			if p.Cargo[i] > 0 {
				good := s.gs.Data.Goods[i]
				name := good.Name
				if !good.Legal {
					name = IllegalStyle.Render(name)
				}
				avgCost := ""
				if p.CargoCost[i] > 0 {
					avg := p.CargoCost[i] / p.Cargo[i]
					avgCost = DimStyle.Render(fmt.Sprintf("  (avg %d cr/ea)", avg))
				}
				b.WriteString(fmt.Sprintf("    %-12s %4d%s\n", name, p.Cargo[i], avgCost))
			}
		}
	} else {
		b.WriteString(DimStyle.Render("  Cargo hold empty") + "\n")
	}

	if s.gs.Quests.TribbleQty > 0 {
		b.WriteString(fmt.Sprintf("  Tribbles: %d\n", s.gs.Quests.TribbleQty))
	}

	activeQuests := getActiveQuests(s.gs)
	if len(activeQuests) > 0 {
		b.WriteString("\n" + div + "\n")
		b.WriteString(CyanStyle.Render("  Active Quests") + "\n")
		for _, q := range activeQuests {
			b.WriteString("    " + q + "\n")
		}
	}

	b.WriteString("\n" + DimStyle.Render("  esc to go back"))
	return b.String()
}

func getActiveQuests(gs *game.GameState) []string {
	var quests []string
	type questInfo struct {
		id   game.QuestID
		name string
		hint string
	}
	checks := []questInfo{
		{game.QuestDragonfly, "Dragonfly", "Chase through Baratas, Melina, Regulas, Zalkon"},
		{game.QuestSpaceMonster, "Space Monster", "Destroy at Acamar"},
		{game.QuestScarab, "Scarab", "Find near a wormhole exit"},
		{game.QuestAlienArtifact, "Alien Artifact", "Deliver to a Hi-tech system"},
		{game.QuestJarek, "Ambassador Jarek", "Transport to Devidia"},
		{game.QuestJapori, "Japori Disease", "Deliver 10 medicine"},
		{game.QuestGemulon, "Gemulon Invasion", "Warn Gemulon (timed!)"},
		{game.QuestFehler, "Dr. Fehler", "Stop experiment at Daled (timed!)"},
		{game.QuestWild, "Jonathan Wild", "Smuggle to Kravat"},
		{game.QuestReactor, "Reactor Delivery", "Deliver to Nix (fuel leak!)"},
	}
	for _, c := range checks {
		state := gs.Quests.States[c.id]
		if state == game.QuestActive || state == game.QuestAvailable {
			quests = append(quests, c.name+" - "+c.hint)
		}
	}
	return quests
}
