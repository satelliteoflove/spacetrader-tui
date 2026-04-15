package screens

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/the4ofus/spacetrader-tui/internal/formula"
	"github.com/the4ofus/spacetrader-tui/internal/game"
	"github.com/the4ofus/spacetrader-tui/internal/gamedata"
)

type activeQuest struct {
	id     game.QuestID
	name   string
	desc   string
	destID int
}

type StatusScreen struct {
	gs      *game.GameState
	quests  []activeQuest
	cursor  int
	message string
}

func NewStatusScreen(gs *game.GameState) *StatusScreen {
	quests := buildActiveQuests(gs)
	return &StatusScreen{gs: gs, quests: quests}
}

func (s *StatusScreen) Init() tea.Cmd { return nil }

func (s *StatusScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if len(s.quests) == 0 {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			if key.Matches(msg, Keys.Back) {
				return s, func() tea.Msg { return NavigateMsg{Screen: ScreenSystem} }
			}
		}
		return s, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, Keys.Up):
			s.cursor = wrapCursor(s.cursor, -1, len(s.quests))
			s.message = ""
		case key.Matches(msg, Keys.Down):
			s.cursor = wrapCursor(s.cursor, 1, len(s.quests))
			s.message = ""
		case msg.String() == "m":
			q := s.quests[s.cursor]
			if q.destID >= 0 {
				destIdx := q.destID
				return s, func() tea.Msg { return NavigateMsg{Screen: ScreenGalacticChart, SelectedSystem: destIdx} }
			}
			s.message = DimStyle.Render("No fixed destination for this quest.")
		case msg.String() == "p":
			q := s.quests[s.cursor]
			if q.destID >= 0 {
				destIdx := q.destID
				return s, func() tea.Msg { return NavigateMsg{Screen: ScreenRoutePlanner, SelectedSystem: destIdx} }
			}
			s.message = DimStyle.Render("No fixed destination for this quest.")
		case key.Matches(msg, Keys.Back):
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
	dp := &game.GameDataProvider{Data: s.gs.Data}
	b.WriteString(fmt.Sprintf("  Net worth: %d cr\n", p.Worth(dp)))
	b.WriteString(fmt.Sprintf("  Credits: %d\n", p.Credits))
	if p.LoanBalance > 0 {
		b.WriteString(DangerStyle.Render(fmt.Sprintf("  Debt: %d", p.LoanBalance)) + "\n")
	}

	b.WriteString("  Skills:\n")
	for i, name := range formula.SkillNames {
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
			wage := m.Wage()
			if m.IsQuest {
				b.WriteString(fmt.Sprintf(" %s (free)", m.Name))
			} else {
				b.WriteString(fmt.Sprintf(" %s (%d/d)", m.Name, wage))
			}
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
					name = IllegalStyle.Render("! " + name)
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
	if s.gs.Quests.HasSingularity {
		b.WriteString(SuccessStyle.Render("  Portable Singularity: READY") + "\n")
	}
	if s.gs.Quests.States[game.QuestReactor] == game.QuestActive {
		status := s.gs.Quests.Progress[game.QuestReactor]
		b.WriteString(DangerStyle.Render(fmt.Sprintf("  Reactor: %d/20 (meltdown at 21!)", status)) + "\n")
	}
	if s.gs.Quests.FabricRipDays > 0 {
		b.WriteString(DangerStyle.Render(fmt.Sprintf("  Fabric Rip: %d days remaining", s.gs.Quests.FabricRipDays)) + "\n")
	}

	if len(s.quests) > 0 {
		b.WriteString("\n" + div + "\n")
		b.WriteString(CyanStyle.Render("  Active Quests") + "\n")
		for i, q := range s.quests {
			line := renderQuestLine(q.name + " - " + q.desc)
			if i == s.cursor {
				b.WriteString(SelectedStyle.Render("  > ") + line + "\n")
			} else {
				b.WriteString("    " + line + "\n")
			}
		}
		if s.message != "" {
			b.WriteString("  " + s.message + "\n")
		}
		b.WriteString("\n" + DimStyle.Render("  j/k select quest  m map  p plan route  esc back"))
	} else {
		b.WriteString("\n" + DimStyle.Render("  esc back"))
	}

	return b.String()
}

func buildActiveQuests(gs *game.GameState) []activeQuest {
	type questInfo struct {
		id   game.QuestID
		name string
	}
	checks := []questInfo{
		{game.QuestDragonfly, "Dragonfly"},
		{game.QuestSpaceMonster, "Space Monster"},
		{game.QuestScarab, "Scarab"},
		{game.QuestAlienArtifact, "Alien Artifact"},
		{game.QuestJarek, "Ambassador Jarek"},
		{game.QuestJapori, "Japori Disease"},
		{game.QuestGemulon, "Gemulon Invasion"},
		{game.QuestFehler, "Dr. Fehler"},
		{game.QuestWild, "Jonathan Wild"},
		{game.QuestReactor, "Reactor Delivery"},
		{game.QuestMoonForSale, "Moon For Sale"},
	}
	var quests []activeQuest
	for _, c := range checks {
		state := gs.Quests.States[c.id]
		if state == game.QuestActive || state == game.QuestAvailable {
			quests = append(quests, activeQuest{
				id:     c.id,
				name:   c.name,
				desc:   gs.QuestDescription(c.id),
				destID: gs.QuestDestination(c.id),
			})
		}
	}
	return quests
}

func renderQuestLine(line string) string {
	var result strings.Builder
	for len(line) > 0 {
		dimIdx := strings.Index(line, "\x00dim\x00")
		nextIdx := strings.Index(line, "\x00next\x00")
		markerIdx := -1
		markerTag := ""
		markerLen := 0
		if dimIdx >= 0 && (nextIdx < 0 || dimIdx < nextIdx) {
			markerIdx = dimIdx
			markerTag = "dim"
			markerLen = len("\x00dim\x00")
		} else if nextIdx >= 0 {
			markerIdx = nextIdx
			markerTag = "next"
			markerLen = len("\x00next\x00")
		}
		if markerIdx < 0 {
			result.WriteString(line)
			break
		}
		result.WriteString(line[:markerIdx])
		line = line[markerIdx+markerLen:]
		endIdx := strings.Index(line, "\x00"+markerTag+"\x00")
		if endIdx < 0 {
			result.WriteString(line)
			break
		}
		content := line[:endIdx]
		line = line[endIdx+markerLen:]
		switch markerTag {
		case "dim":
			result.WriteString(DimStyle.Render(content))
		case "next":
			result.WriteString(SelectedStyle.Render(content))
		}
	}
	return result.String()
}
