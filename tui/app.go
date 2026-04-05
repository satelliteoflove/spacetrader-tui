package tui

import (
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
	case screens.NavigateMsg:
		m.systemHubCursor = msg.RestoreCursor
		return m.navigate(msg.Screen)
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

func (m Model) navigate(screen screens.ScreenType) (tea.Model, tea.Cmd) {
	var s tea.Model
	switch screen {
	case screens.ScreenTitle:
		s = screens.NewTitleScreen()
	case screens.ScreenNewGame:
		s = screens.NewNewGameScreen()
	case screens.ScreenSystem:
		s = screens.NewSystemScreenWithCursor(m.gs, m.systemHubCursor)
	case screens.ScreenMarket:
		s = screens.NewMarketScreen(m.gs)
	case screens.ScreenChart:
		s = screens.NewChartScreen(m.gs)
	case screens.ScreenShipyard:
		s = screens.NewShipyardScreen(m.gs)
	case screens.ScreenBank:
		s = screens.NewBankScreen(m.gs)
	case screens.ScreenPersonnel:
		s = screens.NewPersonnelScreen(m.gs)
	case screens.ScreenGalacticChart:
		s = screens.NewGalacticChartScreen(m.gs)
	case screens.ScreenGalacticList:
		s = screens.NewGalacticListScreen(m.gs)
	case screens.ScreenStatus:
		s = screens.NewStatusScreen(m.gs)
	case screens.ScreenSave:
		s = screens.NewSaveScreen(m.gs)
	case screens.ScreenGameOver:
		s = screens.NewGameOverScreen(m.gs)
	default:
		return m, nil
	}
	m.screen = s
	return m, s.Init()
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

	frame := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("8")).
		Width(maxW).
		Padding(0, 1)

	rendered := frame.Render(content)
	return lipgloss.Place(w, h, lipgloss.Center, lipgloss.Top, rendered)
}
