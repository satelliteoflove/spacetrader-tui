package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/the4ofus/spacetrader-tui/internal/encounter"
	"github.com/the4ofus/spacetrader-tui/internal/game"
	"github.com/the4ofus/spacetrader-tui/internal/gamedata"
	"github.com/the4ofus/spacetrader-tui/tui/screens"
)

type Model struct {
	gs              *game.GameState
	data            *gamedata.GameData
	screen          tea.Model
	width           int
	height          int
	systemHubCursor int
}

func NewModel(data *gamedata.GameData) Model {
	return Model{
		data:   data,
		screen: screens.NewTitleScreen(),
	}
}

func (m Model) Init() tea.Cmd {
	return m.screen.Init()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		var cmd tea.Cmd
		m.screen, cmd = m.screen.Update(msg)
		return m, cmd
	case screens.NavigateMsg:
		m.systemHubCursor = msg.RestoreCursor
		return m.navigate(msg)
	case screens.LoadGameMsg:
		path, err := game.DefaultSavePath()
		if err != nil {
			return m, nil
		}
		gs, err := game.Load(path, m.data)
		if err != nil {
			return m, nil
		}
		m.gs = gs
		s := screens.NewSystemScreen(m.gs)
		m.screen = s
		return m, s.Init()
	case screens.StartGameMsg:
		m.gs = game.NewGame(m.data, msg.Name, msg.Skills, msg.Difficulty)
		s := screens.NewSystemScreen(m.gs)
		m.screen = s
		return m, s.Init()
	case screens.TravelMsg:
		enc := encounter.Generate(m.gs)
		if enc != nil {
			s := screens.NewEncounterScreen(m.gs, enc)
			m.screen = s
			return m, s.Init()
		}
		m.systemHubCursor = 1
		return m.arriveAtSystem()
	case screens.EncounterDoneMsg:
		if m.gs.EndStatus == game.StatusDead {
			s := screens.NewGameOverScreen(m.gs)
			m.screen = s
			return m, s.Init()
		}
		m.systemHubCursor = 1
		return m.arriveAtSystem()
	}

	var cmd tea.Cmd
	m.screen, cmd = m.screen.Update(msg)
	return m, cmd
}

func (m Model) arriveAtSystem() (tea.Model, tea.Cmd) {
	events := game.CheckQuestsOnArrival(m.gs)
	if len(events) > 0 {
		s := screens.NewQuestEventScreen(m.gs, events)
		m.screen = s
		return m, s.Init()
	}
	s := screens.NewSystemScreenWithCursor(m.gs, m.systemHubCursor)
	m.screen = s
	return m, s.Init()
}

func (m Model) navigate(msg screens.NavigateMsg) (tea.Model, tea.Cmd) {
	var s tea.Model
	switch msg.Screen {
	case screens.ScreenTitle:
		s = screens.NewTitleScreen()
	case screens.ScreenNewGame:
		s = screens.NewNewGameScreen()
	case screens.ScreenSystem:
		s = screens.NewSystemScreenWithCursor(m.gs, m.systemHubCursor)
	case screens.ScreenMarket:
		s = screens.NewMarketScreen(m.gs)
	case screens.ScreenShipyard:
		s = screens.NewShipyardScreen(m.gs)
	case screens.ScreenBank:
		s = screens.NewBankScreen(m.gs)
	case screens.ScreenPersonnel:
		s = screens.NewPersonnelScreen(m.gs)
	case screens.ScreenGalacticChart:
		s = screens.NewGalacticChartScreenWithSelection(m.gs, msg.SelectedSystem)
	case screens.ScreenGalacticList:
		s = screens.NewGalacticListScreenWithSelection(m.gs, msg.SelectedSystem)
	case screens.ScreenStatus:
		s = screens.NewStatusScreen(m.gs)
	case screens.ScreenSave:
		s = screens.NewSaveScreen(m.gs)
	case screens.ScreenGameOver:
		s = screens.NewGameOverScreen(m.gs)
	case screens.ScreenGuide:
		s = screens.NewGuideScreen()
	case screens.ScreenNews:
		s = screens.NewNewsScreen(m.gs)
	default:
		return m, nil
	}
	m.screen = s
	return m, s.Init()
}

var (
	statusBarStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("15")).Bold(true)
	statusDimStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	statusDangerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true)
)

func (m Model) statusBar(width int) string {
	if m.gs == nil {
		return ""
	}
	shipDef := m.gs.PlayerShipDef()
	cargo := m.gs.Player.TotalCargo()
	dp := &game.GameDataProvider{Data: m.gs.Data}
	cap := m.gs.Player.CargoCapacity(dp)

	hullPct := m.gs.Player.Ship.Hull * 100 / shipDef.Hull
	hullStr := fmt.Sprintf("Hull:%d%%", hullPct)
	if hullPct < 50 {
		hullStr = statusDangerStyle.Render(hullStr)
	} else {
		hullStr = statusBarStyle.Render(hullStr)
	}

	parts := []string{
		statusBarStyle.Render(fmt.Sprintf("Cr:%d", m.gs.Player.Credits)),
		statusBarStyle.Render(fmt.Sprintf("Cargo:%d/%d", cargo, cap)),
		hullStr,
		statusBarStyle.Render(fmt.Sprintf("Fuel:%d/%d", m.gs.Player.Ship.Fuel, shipDef.Range)),
		statusDimStyle.Render(fmt.Sprintf("Day %d", m.gs.Day)),
	}

	if m.gs.Player.LoanBalance > 0 {
		parts = append(parts, statusDangerStyle.Render(fmt.Sprintf("Debt:%d", m.gs.Player.LoanBalance)))
	}

	return "\n  " + strings.Join(parts, statusDimStyle.Render(" | "))
}

func (m Model) View() string {
	content := m.screen.View()

	w := m.width
	h := m.height
	if w == 0 {
		w = 80
	}
	if h == 0 {
		h = 24
	}

	maxW := 80
	if w-2 < maxW {
		maxW = w - 2
	}

	content += m.statusBar(maxW)

	maxH := h - 2
	if maxH < 10 {
		maxH = 10
	}

	frame := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("8")).
		Width(maxW).
		Height(maxH).
		Padding(0, 1)

	rendered := frame.Render(content)
	return lipgloss.Place(w, h, lipgloss.Center, lipgloss.Top, rendered)
}
