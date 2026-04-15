package screens

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

type guideTab int

const (
	guideTrading guideTab = iota
	guideTech
	guideGov
	guideSpecialty
	guideNavigation
	guideShipCrew
	numGuideTabs
)

type GuideScreen struct {
	tab    guideTab
	scroll int
}

func NewGuideScreen() *GuideScreen {
	return &GuideScreen{}
}

func (s *GuideScreen) Init() tea.Cmd { return nil }

func (s *GuideScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case msg.String() == "1":
			s.tab = guideTrading
			s.scroll = 0
		case msg.String() == "2":
			s.tab = guideTech
			s.scroll = 0
		case msg.String() == "3":
			s.tab = guideGov
			s.scroll = 0
		case msg.String() == "4":
			s.tab = guideSpecialty
			s.scroll = 0
		case msg.String() == "5":
			s.tab = guideNavigation
			s.scroll = 0
		case msg.String() == "6":
			s.tab = guideShipCrew
			s.scroll = 0
		case key.Matches(msg, Keys.Up):
			if s.scroll > 0 {
				s.scroll--
			}
		case key.Matches(msg, Keys.Down):
			s.scroll++
		case key.Matches(msg, Keys.Back):
			return s, func() tea.Msg { return NavigateMsg{Screen: ScreenSystem} }
		}
	}
	return s, nil
}

func (s *GuideScreen) View() string {
	var b strings.Builder

	b.WriteString(HeaderStyle.Render("  TRADER'S GUIDE  ") + "\n")

	tabs := []string{"[1] Trading", "[2] Tech", "[3] Gov", "[4] Specialty", "[5] Navigation", "[6] Ship/Crew"}
	b.WriteString("  ")
	for i, t := range tabs {
		if guideTab(i) == s.tab {
			b.WriteString(SelectedStyle.Render(t) + "  ")
		} else {
			b.WriteString(DimStyle.Render(t) + "  ")
		}
	}
	b.WriteString("\n\n")

	var content string
	switch s.tab {
	case guideTrading:
		content = guideTradingContent()
	case guideTech:
		content = guideTechContent()
	case guideGov:
		content = guideGovContent()
	case guideSpecialty:
		content = guideSpecialtyContent()
	case guideNavigation:
		content = guideNavigationContent()
	case guideShipCrew:
		content = guideShipCrewContent()
	}

	lines := strings.Split(content, "\n")
	if s.scroll >= len(lines) {
		s.scroll = len(lines) - 1
	}
	if s.scroll < 0 {
		s.scroll = 0
	}
	pageSize := 18
	end := s.scroll + pageSize
	if end > len(lines) {
		end = len(lines)
	}
	for _, line := range lines[s.scroll:end] {
		b.WriteString(line + "\n")
	}

	if end < len(lines) {
		b.WriteString(DimStyle.Render(fmt.Sprintf("  -- %d more lines (j to scroll) --", len(lines)-end)) + "\n")
	}

	b.WriteString("\n" + DimStyle.Render("  1-6 tabs, j/k scroll, esc back"))
	return b.String()
}

func guideTradingContent() string {
	return fmt.Sprintf(`  %s
  Buy goods where they are cheap, sell where expensive.
  The market screen shows trend arrows to help:

    %s  very cheap -- great time to buy
    %s   cheap -- good buy
    %s   average price
    %s   expensive -- good time to sell
    %s  very expensive -- great time to sell

  %s
  Each good has a tech range. Low-tech systems produce
  basic goods (Water, Furs, Food) cheaply. High-tech
  systems produce advanced goods (Machines, Robots).

  Buy low-tech goods at low-tech worlds, sell at
  high-tech worlds. Buy high-tech goods at high-tech
  worlds, sell at low-tech worlds.

  %s
  Firearms and Narcotics are illegal. Police may
  inspect your cargo and confiscate contraband.
  Dictatorship, Fascist, and Military governments
  mark up illegal goods +50%% -- but this also means
  you can sell them for more elsewhere.

  %s
  Your Trader skill reduces prices by 1%% per point
  (max 10%%). Crew members with high Trader skill
  also contribute.

  %s
  The Navigation list shows a GOODS column with 10
  letters, one per commodity in fixed order:

    %s

  Letter colors indicate price vs galactic average:
    %s = below average (cheap to buy)
    %s = near average
    %s = above average (expensive)
    %s = not available at this system

  This requires trade info (see Navigation tab).

  %s
  The Portfolio screen (from system menu) tracks your
  credits and net worth over time as a chart. It
  updates each time you arrive at a new system.`,
		CyanStyle.Render("BASICS"),
		SuccessStyle.Render("<<"),
		SuccessStyle.Render("<"),
		DimStyle.Render("="),
		DangerStyle.Render(">"),
		DangerStyle.Render(">>"),
		CyanStyle.Render("TECH AND PRICES"),
		CyanStyle.Render("ILLEGAL GOODS"),
		CyanStyle.Render("TRADER SKILL"),
		CyanStyle.Render("GOODS COLUMN"),
		DimStyle.Render("W=Water U=Furs F=Food O=Ore G=Games A=Firearms D=Medicine C=Machines N=Narcotics R=Robots"),
		SuccessStyle.Render("green"),
		"white",
		DangerStyle.Render("red"),
		DimStyle.Render("dim"),
		CyanStyle.Render("PORTFOLIO"))
}

func guideTechContent() string {
	return fmt.Sprintf(`  %s
  Each system has a tech level that determines which
  goods are available for trade there.

  %s
  %-16s %s
  %-16s %s
  %-16s %s
  %-16s %s
  %-16s %s
  %-16s %s
  %-16s %s
  %-16s %s

  %s
  Higher tech levels unlock more advanced (and
  more expensive) trade goods:

  %-16s %s
  %-16s %s
  %-16s %s
  %-16s %s
  %-16s %s
  %-16s %s
  %-16s %s
  %-16s %s
  %-16s %s
  %-16s %s`,
		CyanStyle.Render("TECHNOLOGY LEVELS"),
		DimStyle.Render("  Level            Description"),
		"  Pre-ag", "Primitive, only water and furs",
		"  Agricultural", "Farming begins, food is cheap here",
		"  Medieval", "Ore production begins, cheap here",
		"  Renaissance", "Games and firearms appear",
		"  Early Ind", "Medicine and machines produced",
		"  Industrial", "Mass production, narcotics appear",
		"  Post-ind", "Robots produced, advanced equipment",
		"  Hi-tech", "All goods available, basics cost more",
		CyanStyle.Render("GOODS BY TECH LEVEL"),
		"  Water", "Pre-ag through Hi-tech",
		"  Furs", "Pre-ag through Hi-tech",
		"  Food", "Agricultural and up",
		"  Ore", "Medieval and up",
		"  Games", "Renaissance and up",
		"  Firearms", "Renaissance and up (illegal)",
		"  Medicine", "Early Industrial and up",
		"  Machines", "Early Industrial and up",
		"  Narcotics", "Industrial and up (illegal)",
		"  Robots", "Post-industrial and up")
}

func guideGovContent() string {
	return fmt.Sprintf(`  %s
  Government type affects two things: encounter
  danger and illegal goods pricing.

  %s
  %s  High pirates, no police
    Anarchy, Feudal State

  %s  Heavy policing, few pirates
    Military State, Fascist State

  %s  Moderate balance
    Democracy, Corporate State, Technocracy,
    Cybernetic State, Monarchy, Confederacy

  %s  Low activity overall
    Pacifist State, State of Satori

  %s  More police than pirates
    Communist State, Socialist State, Theocracy

  %s
  Dictatorship, Fascist State, and Military State
  mark up illegal goods (Firearms, Narcotics) by
  +50%%. This makes them expensive to buy there --
  but also means these goods sell well at other
  government types.`,
		CyanStyle.Render("GOVERNMENT TYPES"),
		CyanStyle.Render("ENCOUNTER DANGER"),
		DangerStyle.Render("  Lawless"),
		SuccessStyle.Render("  Heavily Policed"),
		NormalStyle.Render("  Moderate"),
		DimStyle.Render("  Peaceful"),
		NormalStyle.Render("  Controlled"),
		CyanStyle.Render("ILLEGAL GOODS MARKUP"))
}

func guideSpecialtyContent() string {
	return fmt.Sprintf(`  %s
  Some systems have a specialty that affects local
  prices. These are shown in the system detail screen
  (press d from Navigation):

    %s  Abundance -- goods are cheaper here
    %s  Scarcity -- goods cost more here

  %s
  %s  Water cheap here
  %s  Minerals cheap here
  %s  Furs cheap here (also Furs decrease event)
  %s  Food cheap here
  %s  Medicine cheap here
  %s  Machines/labor cheap here

  %s
  %s  Water expensive here
  %s  General poverty, higher prices
  %s  Furs expensive here (lifeless world)
  %s  Food expensive here
  %s  Medicine expensive here
  %s  Labor shortage, higher costs

  %s
  Buy at %s systems, sell at %s systems.
  Example: buy Water at a %s world,
  sell it at a %s world for +50%% profit.`,
		CyanStyle.Render("SYSTEM SPECIALTIES"),
		SuccessStyle.Render("+"),
		DangerStyle.Render("-"),
		CyanStyle.Render("ABUNDANCE (+) -- BUY HERE"),
		SuccessStyle.Render("  +Water"),
		SuccessStyle.Render("  +Minerals"),
		SuccessStyle.Render("  +Fauna"),
		SuccessStyle.Render("  +Soil"),
		SuccessStyle.Render("  +Good med"),
		SuccessStyle.Render("  +Robots"),
		CyanStyle.Render("SCARCITY (-) -- SELLS HIGH HERE"),
		DangerStyle.Render("  -Desert"),
		DangerStyle.Render("  -Poor"),
		DangerStyle.Render("  -Lifeless"),
		DangerStyle.Render("  -Poor soil"),
		DangerStyle.Render("  -Poor med"),
		DangerStyle.Render("  -Low labor"),
		CyanStyle.Render("TRADE STRATEGY"),
		SuccessStyle.Render("+"),
		DangerStyle.Render("-"),
		SuccessStyle.Render("+Water"),
		DangerStyle.Render("-Desert"))
}

func guideNavigationContent() string {
	return fmt.Sprintf(`  %s
  The Navigation list shows all systems with distance,
  tech level, government, and a GOODS column. Press
  enter to travel, d for system details.

  %s
  Trade info is gathered automatically when you visit
  a system. You can also purchase info for systems
  within 15 parsecs by pressing i (costs 100 cr).

  Trade info goes stale after 5 days. Stale data is
  shown in grey. Visit or re-purchase to refresh.

  %s
  The system detail screen (d) shows full info about
  a system: tech level, government, resources, events,
  and estimated commodity prices from your trade info.

  Prices are color-coded vs the galactic average:
    %s = below average (buy opportunity)
    %s = above average (sell opportunity)

  If you are carrying cargo, profit/loss per unit is
  shown based on your cost basis.

  %s
  The Route Planner (p from Navigation) finds multi-hop
  paths to distant systems. It shows fuel costs, refuel
  expenses, and trade opportunities at each hop.

  Trade estimates in the route planner require trade
  info for both systems in each hop. If you are missing
  info, it will tell you.

  Press r to set a route as active. Your active route
  is shown on the system menu for quick access.

  %s
  Wormholes connect distant systems instantly. They
  appear on the Navigation list and galactic map.
  Transit costs credits (based on ship size) but
  no fuel. Press w from Navigation when available.

  %s
  The Galaxy News screen shows recent events that
  affect commodity prices. Watch for headlines about
  your destination -- they can mean big profits or
  losses depending on what you are carrying.`,
		CyanStyle.Render("NAVIGATION"),
		CyanStyle.Render("TRADE INFO"),
		CyanStyle.Render("SYSTEM DETAILS"),
		SuccessStyle.Render("green"),
		DangerStyle.Render("red"),
		CyanStyle.Render("ROUTE PLANNER"),
		CyanStyle.Render("WORMHOLES"),
		CyanStyle.Render("GALAXY NEWS"))
}

func guideShipCrewContent() string {
	return fmt.Sprintf(`  %s
  Visit the Shipyard to buy a new ship, equipment,
  repairs, and fuel. Trading in your ship gives you
  75%% of its value toward the new one.

  %s
  Equipment resells for 75%% of its purchase price.
  Use the Sell tab in the Shipyard to offload gear
  you no longer need.

  %s
  Weapons: Pulse Laser, Beam Laser, Military Laser
  Shields: Energy, Reflective
  Gadgets: Cargo Bays, Fuel Compactor, Nav System,
           Auto-Repair, Cloaking Device

  Better equipment requires higher tech level systems.

  %s
  Hire mercenaries at the Personnel screen. They add
  their skills to yours, boosting combat, trading,
  piloting, and repair. Each merc has a daily wage
  based on their total skill level.

  If you cannot pay wages, all crew is dismissed.

  %s
  Buy an Escape Pod to survive ship destruction.
  Without one, losing your ship means game over.

  With an Escape Pod, you can buy Insurance. If your
  ship is destroyed, insurance pays 75%% of the ship
  and equipment value. The premium decreases the
  longer you go without a claim.

  %s
  Combat uses your Fighter skill for hit chance and
  your Pilot skill for evasion. All equipped weapons
  fire each round. Shields absorb damage before hull.
  Engineer skill provides passive hull repair each day.

  %s
  The Bank offers loans up to 10%% of your net worth
  (1,000 to 25,000 cr). Interest is 10%% of the
  balance per day. Pay off debts quickly -- interest
  compounds fast.`,
		CyanStyle.Render("SHIPYARD"),
		CyanStyle.Render("SELLING EQUIPMENT"),
		CyanStyle.Render("EQUIPMENT TYPES"),
		CyanStyle.Render("CREW"),
		CyanStyle.Render("ESCAPE POD & INSURANCE"),
		CyanStyle.Render("COMBAT"),
		CyanStyle.Render("BANK & LOANS"))
}
