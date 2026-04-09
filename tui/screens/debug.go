package screens

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/the4ofus/spacetrader-tui/internal/game"
	"github.com/the4ofus/spacetrader-tui/internal/gamedata"
)

const debugMaxLines = 20

var questNameMap = map[string]game.QuestID{
	"dragonfly": game.QuestDragonfly,
	"monster":   game.QuestSpaceMonster,
	"scarab":    game.QuestScarab,
	"artifact":  game.QuestAlienArtifact,
	"jarek":     game.QuestJarek,
	"japori":    game.QuestJapori,
	"gemulon":   game.QuestGemulon,
	"fehler":    game.QuestFehler,
	"wild":      game.QuestWild,
	"reactor":   game.QuestReactor,
	"tribbles":  game.QuestTribbles,
	"skill":     game.QuestSkillIncrease,
	"erase":     game.QuestEraseRecord,
	"cargo":     game.QuestCargoForSale,
	"lottery":   game.QuestLotteryWinner,
	"moon":      game.QuestMoonForSale,
	"fabricrip": game.QuestFabricRip,
}

var questStateMap = map[string]game.QuestState{
	"unavail":  game.QuestUnavailable,
	"avail":    game.QuestAvailable,
	"active":   game.QuestActive,
	"complete": game.QuestComplete,
}

type DebugScreen struct {
	gs     *game.GameState
	input  textinput.Model
	output []string
}

func NewDebugScreen(gs *game.GameState) *DebugScreen {
	ti := textinput.New()
	ti.Placeholder = "type a command (help for list)"
	ti.Focus()
	ti.CharLimit = 80

	return &DebugScreen{
		gs:     gs,
		input:  ti,
		output: []string{"Debug console ready. Type 'help' for commands."},
	}
}

func (s *DebugScreen) Init() tea.Cmd {
	return textinput.Blink
}

func (s *DebugScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return s, func() tea.Msg { return NavigateMsg{Screen: ScreenSystem} }
		case "enter":
			cmd := strings.TrimSpace(s.input.Value())
			if cmd != "" {
				result := s.execute(cmd)
				s.output = append(s.output, "> "+cmd)
				s.output = append(s.output, result)
			}
			s.input.SetValue("")
			return s, nil
		}
	}
	var cmd tea.Cmd
	s.input, cmd = s.input.Update(msg)
	return s, cmd
}

func (s *DebugScreen) execute(cmd string) string {
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return ""
	}
	verb := strings.ToLower(parts[0])
	args := parts[1:]

	switch verb {
	case "help":
		return s.cmdHelp()
	case "day":
		return s.cmdSetInt(args, "day", func(n int) { s.gs.Day = n })
	case "credits":
		return s.cmdSetInt(args, "credits", func(n int) { s.gs.Player.Credits = n })
	case "police":
		return s.cmdSetInt(args, "police record", func(n int) { s.gs.Player.PoliceRecord = n })
	case "rep":
		return s.cmdSetInt(args, "reputation", func(n int) { s.gs.Player.Reputation = n })
	case "fuel":
		return s.cmdSetInt(args, "fuel", func(n int) { s.gs.Player.Ship.Fuel = n })
	case "hull":
		return s.cmdSetInt(args, "hull", func(n int) { s.gs.Player.Ship.Hull = n })
	case "goto":
		return s.cmdGoto(args)
	case "quest":
		return s.cmdQuest(args)
	case "questprog":
		return s.cmdQuestProg(args)
	case "cargo":
		return s.cmdCargo(args)
	case "equip":
		return s.cmdEquip(args)
	case "tribbles":
		return s.cmdSetInt(args, "tribbles", func(n int) { s.gs.Quests.TribbleQty = n })
	case "singularity":
		s.gs.Quests.HasSingularity = !s.gs.Quests.HasSingularity
		return fmt.Sprintf("Singularity: %v", s.gs.Quests.HasSingularity)
	case "escapepod":
		s.gs.Player.HasEscapePod = !s.gs.Player.HasEscapePod
		return fmt.Sprintf("Escape pod: %v", s.gs.Player.HasEscapePod)
	case "monsterhull":
		return s.cmdSetInt(args, "monster hull", func(n int) { s.gs.Quests.MonsterHull = n })
	case "fabricrip":
		return s.cmdSetInt(args, "fabric rip days", func(n int) { s.gs.Quests.FabricRipDays = n })
	default:
		return fmt.Sprintf("Unknown command: %s", verb)
	}
}

func (s *DebugScreen) cmdHelp() string {
	return strings.Join([]string{
		"day <N>          -- set game day",
		"credits <N>      -- set credits",
		"police <N>       -- set police record",
		"rep <N>          -- set reputation",
		"fuel <N>         -- set ship fuel",
		"hull <N>         -- set ship hull",
		"goto <system>    -- teleport to system",
		"quest <name> <state> -- set quest state",
		"  states: unavail/avail/active/complete",
		"questprog <name> <N> -- set quest progress",
		"cargo <good> <N> -- set cargo amount",
		"equip <name>     -- add equipment",
		"tribbles <N>     -- set tribble count",
		"singularity      -- toggle singularity",
		"escapepod        -- toggle escape pod",
		"monsterhull <N>  -- set monster hull",
		"fabricrip <N>    -- set fabric rip days",
	}, "\n")
}

func (s *DebugScreen) cmdSetInt(args []string, label string, setter func(int)) string {
	if len(args) < 1 {
		return fmt.Sprintf("Usage: %s <N>", label)
	}
	n, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Sprintf("Invalid number: %s", args[0])
	}
	setter(n)
	return fmt.Sprintf("Set %s to %d", label, n)
}

func (s *DebugScreen) cmdGoto(args []string) string {
	if len(args) < 1 {
		return "Usage: goto <system_name>"
	}
	name := strings.Join(args, " ")
	nameLower := strings.ToLower(name)
	for i, sys := range s.gs.Data.Systems {
		if strings.ToLower(sys.Name) == nameLower {
			s.gs.CurrentSystemID = i
			s.gs.Systems[i].Visited = true
			game.GenerateEvents(s.gs)
			game.RefreshSystemPrices(s.gs, i)
			return fmt.Sprintf("Teleported to %s", sys.Name)
		}
	}
	return fmt.Sprintf("System not found: %s", name)
}

func (s *DebugScreen) cmdQuest(args []string) string {
	if len(args) < 2 {
		return "Usage: quest <name> <state>"
	}
	qid, ok := s.matchQuest(args[0])
	if !ok {
		return fmt.Sprintf("Unknown quest: %s", args[0])
	}
	stateName := strings.ToLower(args[1])
	state, ok := questStateMap[stateName]
	if !ok {
		return fmt.Sprintf("Unknown state: %s (use unavail/avail/active/complete)", args[1])
	}
	s.gs.SetQuestState(qid, state)
	return fmt.Sprintf("Set quest %s to %s", args[0], stateName)
}

func (s *DebugScreen) cmdQuestProg(args []string) string {
	if len(args) < 2 {
		return "Usage: questprog <name> <N>"
	}
	qid, ok := s.matchQuest(args[0])
	if !ok {
		return fmt.Sprintf("Unknown quest: %s", args[0])
	}
	n, err := strconv.Atoi(args[1])
	if err != nil {
		return fmt.Sprintf("Invalid number: %s", args[1])
	}
	s.gs.SetQuestProgress(qid, n)
	return fmt.Sprintf("Set quest %s progress to %d", args[0], n)
}

func (s *DebugScreen) matchQuest(name string) (game.QuestID, bool) {
	nameLower := strings.ToLower(name)
	if qid, ok := questNameMap[nameLower]; ok {
		return qid, true
	}
	for key, qid := range questNameMap {
		if strings.Contains(key, nameLower) {
			return qid, true
		}
	}
	return 0, false
}

func (s *DebugScreen) cmdCargo(args []string) string {
	if len(args) < 2 {
		return "Usage: cargo <good_name> <N>"
	}
	goodName := strings.ToLower(args[0])
	n, err := strconv.Atoi(args[1])
	if err != nil {
		return fmt.Sprintf("Invalid number: %s", args[1])
	}
	for i, g := range s.gs.Data.Goods {
		if strings.ToLower(g.Name) == goodName {
			s.gs.Player.Cargo[i] = n
			return fmt.Sprintf("Set %s cargo to %d", g.Name, n)
		}
	}
	return fmt.Sprintf("Unknown good: %s", args[0])
}

func (s *DebugScreen) cmdEquip(args []string) string {
	if len(args) < 1 {
		return "Usage: equip <name>"
	}
	name := strings.ToLower(strings.Join(args, " "))
	var found *gamedata.EquipDef
	for i := range s.gs.Data.Equipment {
		eq := &s.gs.Data.Equipment[i]
		if strings.Contains(strings.ToLower(eq.Name), name) {
			found = eq
			break
		}
	}
	if found == nil {
		return fmt.Sprintf("Unknown equipment: %s", strings.Join(args, " "))
	}
	shipDef := s.gs.PlayerShipDef()
	switch found.Category {
	case gamedata.EquipWeapon:
		if len(s.gs.Player.Ship.Weapons) >= shipDef.WeaponSlots {
			return fmt.Sprintf("Weapon slots full (%d/%d)", len(s.gs.Player.Ship.Weapons), shipDef.WeaponSlots)
		}
		s.gs.Player.Ship.Weapons = append(s.gs.Player.Ship.Weapons, found.ID)
	case gamedata.EquipShield:
		if len(s.gs.Player.Ship.Shields) >= shipDef.ShieldSlots {
			return fmt.Sprintf("Shield slots full (%d/%d)", len(s.gs.Player.Ship.Shields), shipDef.ShieldSlots)
		}
		s.gs.Player.Ship.Shields = append(s.gs.Player.Ship.Shields, found.ID)
	case gamedata.EquipGadget:
		if len(s.gs.Player.Ship.Gadgets) >= shipDef.GadgetSlots {
			return fmt.Sprintf("Gadget slots full (%d/%d)", len(s.gs.Player.Ship.Gadgets), shipDef.GadgetSlots)
		}
		s.gs.Player.Ship.Gadgets = append(s.gs.Player.Ship.Gadgets, found.ID)
	}
	return fmt.Sprintf("Added %s", found.Name)
}

func (s *DebugScreen) View() string {
	var b strings.Builder

	b.WriteString(DangerStyle.Render("[DEBUG]") + " Console\n")
	b.WriteString(DimStyle.Render(strings.Repeat("-", 40)) + "\n")

	lines := s.output
	if len(lines) > debugMaxLines {
		lines = lines[len(lines)-debugMaxLines:]
	}
	for _, line := range lines {
		b.WriteString(line + "\n")
	}

	padding := debugMaxLines - len(lines)
	for i := 0; i < padding; i++ {
		b.WriteString("\n")
	}

	b.WriteString(DimStyle.Render(strings.Repeat("-", 40)) + "\n")
	b.WriteString(s.input.View() + "\n")
	b.WriteString(DimStyle.Render("  enter to run, esc to return"))

	return b.String()
}
