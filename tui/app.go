package tui

import (
	"fmt"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/the4ofus/spacetrader-tui/internal/formula"
	"github.com/the4ofus/spacetrader-tui/internal/game"
	"github.com/the4ofus/spacetrader-tui/internal/gamedata"
	"github.com/the4ofus/spacetrader-tui/tui/screens"
)

const tickInterval = 125 * time.Millisecond

func tickCmd() tea.Cmd {
	return tea.Tick(tickInterval, func(t time.Time) tea.Msg {
		return screens.TickMsg{Time: t}
	})
}

func hasSaveFile() bool {
	path, err := game.DefaultSavePath()
	if err != nil {
		return false
	}
	_, err = os.Stat(path)
	return err == nil
}

type Model struct {
	gs              *game.GameState
	data            *gamedata.GameData
	config          game.Config
	screen          tea.Model
	warpScreen      *screens.WarpScreen
	warpDestIdx     int
	width           int
	height          int
	systemHubCursor int
	colorblind      bool
	fadeFrame       int
	pulsePhase      int
}

func NewModel(data *gamedata.GameData, cfg game.Config) Model {
	return Model{
		data:            data,
		config:          cfg,
		colorblind:      cfg.ColorblindMode,
		systemHubCursor: 1,
		screen:          screens.NewTitleScreenWithConfig(cfg.ColorblindMode, hasSaveFile()),
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(m.screen.Init(), tickCmd())
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
		if msg.RestoreCursor > 0 {
			m.systemHubCursor = msg.RestoreCursor
		}
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
		m.fadeFrame = 0
		m.systemHubCursor = 1
		s := screens.NewSystemScreenWithCursor(m.gs, m.systemHubCursor)
		m.screen = s
		return m, s.Init()
	case screens.StartGameMsg:
		m.gs = game.NewGame(m.data, msg.Name, msg.Skills, msg.Difficulty)
		m.fadeFrame = 0
		m.systemHubCursor = 1
		var events []game.QuestEvent
		if msg.Difficulty == gamedata.DiffBeginner {
			events = append(events, game.QuestEvent{
				Title:   "Lottery Winner",
				Message: "You are lucky! While docking at the space port, you receive a\nmessage that you won 1000 credits in a lottery. The prize has\nbeen added to your account.",
			})
		}
		if len(events) > 0 {
			s := screens.NewQuestEventScreen(m.gs, events)
			m.screen = s
			return m, s.Init()
		}
		s := screens.NewSystemScreenWithCursor(m.gs, m.systemHubCursor)
		m.screen = s
		return m, s.Init()
	case screens.TravelMsg:
		cur := m.gs.Data.Systems[m.gs.CurrentSystemID]
		dest := m.gs.Data.Systems[msg.DestIdx]
		dist := formula.Distance(cur.X, cur.Y, dest.X, dest.Y)
		m.fadeFrame = 0
		m.warpDestIdx = msg.DestIdx
		s := screens.NewWarpScreen(m.gs, msg.DestIdx, dest.Name, dist)
		m.warpScreen = s
		m.screen = s
		return m, s.Init()
	case screens.WarpEncounterMsg:
		m.fadeFrame = 0
		s := screens.NewEncounterScreen(m.gs, msg.Encounter)
		m.screen = s
		return m, s.Init()
	case screens.WarpDoneMsg:
		m.warpScreen = nil
		m.systemHubCursor = 1
		return m.arriveAtSystem()
	case screens.EncounterDoneMsg:
		if m.gs.EndStatus == game.StatusDead {
			m.fadeFrame = 0
			m.warpScreen = nil
			s := screens.NewGameOverScreen(m.gs)
			m.screen = s
			return m, s.Init()
		}
		if m.warpScreen != nil {
			m.fadeFrame = 0
			m.screen = m.warpScreen
			m.warpScreen.Update(screens.WarpResumeMsg{})
			return m, nil
		}
		m.systemHubCursor = 1
		return m.arriveAtSystem()
	case screens.TickMsg:
		if m.fadeFrame < screens.AnimFadeDone {
			m.fadeFrame++
		}
		if screens.AnimPulsePhases > 0 {
			m.pulsePhase = (m.pulsePhase + 1) % screens.AnimPulsePhases
		} else {
			m.pulsePhase = 0
		}
		var cmd tea.Cmd
		m.screen, cmd = m.screen.Update(msg)
		return m, tea.Batch(cmd, tickCmd())
	case screens.UpdateSettingsMsg:
		m.config = msg.Config
		if m.colorblind != m.config.ColorblindMode {
			m.colorblind = m.config.ColorblindMode
			screens.InitStyles(m.colorblind)
			InitStatusStyles(m.colorblind)
		}
		screens.ApplyAnimationSettings(m.config)
		return m, nil
	}

	var cmd tea.Cmd
	m.screen, cmd = m.screen.Update(msg)
	return m, cmd
}

func (m Model) arriveAtSystem() (tea.Model, tea.Cmd) {
	m.fadeFrame = 0
	if m.gs.HasActiveRoute && m.gs.CurrentSystemID == m.gs.ActiveRoute {
		m.gs.HasActiveRoute = false
	}
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
		s = screens.NewTitleScreenWithConfig(m.colorblind, hasSaveFile())
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
	case screens.ScreenSettings:
		s = screens.NewSettingsScreen(m.config, m.gs != nil)
	case screens.ScreenRoutePlanner:
		s = screens.NewRoutePlannerScreen(m.gs, msg.SelectedSystem, msg.ReturnScreen)
	case screens.ScreenDebug:
		s = screens.NewDebugScreen(m.gs)
	default:
		return m, nil
	}
	m.fadeFrame = 0
	m.screen = s
	return m, s.Init()
}

var (
	statusBarStyle        lipgloss.Style
	statusDimStyle        lipgloss.Style
	statusDangerStyle     lipgloss.Style
	statusQuestFreshStyle lipgloss.Style
	statusQuestStaleStyle lipgloss.Style
)

func InitStatusStyles(colorblind bool) {
	var dangerColor, successColor lipgloss.TerminalColor
	if colorblind {
		dangerColor = lipgloss.AdaptiveColor{Light: "166", Dark: "208"}
		successColor = lipgloss.AdaptiveColor{Light: "30", Dark: "14"}
	} else {
		dangerColor = lipgloss.AdaptiveColor{Light: "1", Dark: "9"}
		successColor = lipgloss.AdaptiveColor{Light: "2", Dark: "10"}
	}

	statusBarStyle = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "0", Dark: "15"}).Bold(true)
	statusDimStyle = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "242", Dark: "8"})
	statusDangerStyle = lipgloss.NewStyle().Foreground(dangerColor).Bold(true)
	statusQuestFreshStyle = lipgloss.NewStyle().Foreground(successColor).Bold(true)
	statusQuestStaleStyle = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "130", Dark: "11"}).Bold(true)
}

func init() {
	InitStatusStyles(false)
}

var pulseColorsNormal = []string{
	"196", "203", "210", "217", "224", "231",
}
var pulseColorsNormalLight = []string{
	"124", "131", "138", "145", "152", "159",
}
var pulseColorsCB = []string{
	"208", "214", "220", "226", "227", "228",
}
var pulseColorsCBLight = []string{
	"166", "172", "178", "136", "130", "124",
}

func pulseDangerStyle(phase int, colorblind bool) lipgloss.Style {
	if screens.AnimPulsePhases <= 0 {
		if colorblind {
			return lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "166", Dark: "208"}).Bold(true)
		}
		return lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "124", Dark: "196"}).Bold(true)
	}
	darkColors := pulseColorsNormal
	lightColors := pulseColorsNormalLight
	if colorblind {
		darkColors = pulseColorsCB
		lightColors = pulseColorsCBLight
	}
	half := screens.AnimPulsePhases / 2
	idx := phase
	if idx >= half {
		idx = screens.AnimPulsePhases - 1 - idx
	}
	ci := idx * (len(darkColors) - 1) / half
	if ci >= len(darkColors) {
		ci = len(darkColors) - 1
	}
	return lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: lightColors[ci], Dark: darkColors[ci]}).Bold(true)
}

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
		hullStr = pulseDangerStyle(m.pulsePhase, m.colorblind).Render(hullStr)
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
		parts = append(parts, pulseDangerStyle(m.pulsePhase, m.colorblind).Render(fmt.Sprintf("Debt:%d", m.gs.Player.LoanBalance)))
	}

	switch m.gs.QuestUrgency() {
	case game.QuestUrgencyFresh:
		parts = append(parts, statusQuestFreshStyle.Render("Quest:*"))
	case game.QuestUrgencyStale:
		parts = append(parts, statusQuestStaleStyle.Render("Quest:!"))
	case game.QuestUrgencyCritical:
		parts = append(parts, statusDangerStyle.Render("Quest:!!"))
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

	if screens.AnimFadeDone > 0 && m.fadeFrame < screens.AnimFadeDone {
		stripped := screens.StripANSI(content)
		content = screens.FadeStyles[m.fadeFrame].Render(stripped)
		bar := m.statusBar(maxW)
		strippedBar := screens.StripANSI(bar)
		content += screens.FadeStyles[m.fadeFrame].Render(strippedBar)
	} else {
		content += m.statusBar(maxW)
	}

	maxH := h - 2
	if maxH > 45 {
		maxH = 45
	}
	if maxH < 10 {
		maxH = 10
	}

	frame := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.AdaptiveColor{Light: "242", Dark: "8"}).
		Width(maxW).
		Height(maxH).
		Padding(0, 1)

	rendered := frame.Render(content)
	return lipgloss.Place(w, h, lipgloss.Center, lipgloss.Top, rendered)
}
