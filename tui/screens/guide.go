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

	tabs := []string{"[1] Trading", "[2] Tech", "[3] Government", "[4] Specialty"}
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

	b.WriteString("\n" + DimStyle.Render("  1-4 tabs, j/k scroll, esc back"))
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
  also contribute.`,
		CyanStyle.Render("BASICS"),
		SuccessStyle.Render("<<"),
		SuccessStyle.Render("<"),
		DimStyle.Render("="),
		DangerStyle.Render(">"),
		DangerStyle.Render(">>"),
		CyanStyle.Render("TECH AND PRICES"),
		CyanStyle.Render("ILLEGAL GOODS"),
		CyanStyle.Render("TRADER SKILL"))
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
		"  Pre-ag", "Primitive, subsistence farming",
		"  Agricultural", "Organized farming, basic trade",
		"  Medieval", "Metalworking, ore demand rises",
		"  Renaissance", "Early science, games and firearms",
		"  Early Ind", "Factories, medicine and machines",
		"  Industrial", "Mass production, narcotics appear",
		"  Post-ind", "Automation, advanced equipment",
		"  Hi-tech", "Peak technology, all goods available",
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
  prices. On the chart, these are marked with:

    %s  Abundance -- goods are cheaper here
    %s  Scarcity -- goods cost more here
    %s  Neutral/mixed effect

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
		SelectedStyle.Render("~"),
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
