package screens

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/the4ofus/spacetrader-tui/internal/game"
)

var leftRight = key.NewBinding(key.WithKeys("left", "h"))
var rightKey = key.NewBinding(key.WithKeys("right", "l"))

type settingKind int

const (
	settingSpeed settingKind = iota
	settingBool
)

type settingRow struct {
	label string
	kind  settingKind
	speed *game.AnimSpeed
	flag  *bool
}

func (r *settingRow) adjustRight() {
	if r.kind == settingBool {
		*r.flag = !*r.flag
	} else {
		*r.speed = r.speed.Next()
	}
}

func (r *settingRow) adjustLeft() {
	if r.kind == settingBool {
		*r.flag = !*r.flag
	} else {
		*r.speed = r.speed.Prev()
	}
}

type SettingsScreen struct {
	cursor int
	items  []settingRow
	config game.Config
	inGame bool
}

func NewSettingsScreen(cfg game.Config, inGame bool) *SettingsScreen {
	s := &SettingsScreen{config: cfg, inGame: inGame}
	s.items = []settingRow{
		{"Colorblind Mode", settingBool, nil, &s.config.ColorblindMode},
		{"Screen Transitions", settingSpeed, &s.config.TransitionSpeed, nil},
		{"Warp Animation", settingSpeed, &s.config.WarpSpeed, nil},
		{"Encounter Entrance", settingSpeed, &s.config.EncounterEntrance, nil},
		{"Typewriter Text", settingSpeed, &s.config.TypewriterSpeed, nil},
		{"Status Pulse", settingSpeed, &s.config.PulseSpeed, nil},
	}
	return s
}

func (s *SettingsScreen) Init() tea.Cmd { return nil }

func (s *SettingsScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, Keys.Up):
			s.cursor = wrapCursor(s.cursor, -1, len(s.items))
		case key.Matches(msg, Keys.Down):
			s.cursor = wrapCursor(s.cursor, 1, len(s.items))
		case key.Matches(msg, rightKey):
			s.items[s.cursor].adjustRight()
			return s, s.saveAndApply()
		case key.Matches(msg, leftRight):
			s.items[s.cursor].adjustLeft()
			return s, s.saveAndApply()
		case key.Matches(msg, Keys.Back):
			target := ScreenTitle
			if s.inGame {
				target = ScreenSystem
			}
			return s, func() tea.Msg { return NavigateMsg{Screen: target} }
		}
	}
	return s, nil
}

func (s *SettingsScreen) saveAndApply() tea.Cmd {
	game.SaveConfig(s.config)
	cfg := s.config
	return func() tea.Msg { return UpdateSettingsMsg{Config: cfg} }
}

func (s *SettingsScreen) View() string {
	var b strings.Builder

	b.WriteString(TitleStyle.Render("SETTINGS") + "\n")

	speeds := []game.AnimSpeed{game.AnimSlow, game.AnimMedium, game.AnimFast, game.AnimOff}

	for i, item := range s.items {
		cursor := "  "
		style := NormalStyle
		if i == s.cursor {
			cursor = "> "
			style = SelectedStyle
		}

		b.WriteString("  " + style.Render(cursor+item.label) + "\n")
		b.WriteString("    ")

		if item.kind == settingBool {
			for j, label := range []string{"ON", "OFF"} {
				if j > 0 {
					b.WriteString(DimStyle.Render("  "))
				}
				padded := fmt.Sprintf("  %s  ", label)
				isOn := (label == "ON" && *item.flag) || (label == "OFF" && !*item.flag)
				if isOn {
					b.WriteString(SelectedStyle.Render("[" + padded + "]"))
				} else {
					b.WriteString(DimStyle.Render(" " + padded + " "))
				}
			}
		} else {
			for j, spd := range speeds {
				if j > 0 {
					b.WriteString(DimStyle.Render("  "))
				}
				label := fmt.Sprintf(" %s ", spd.String())
				if *item.speed == spd {
					b.WriteString(SelectedStyle.Render("[" + label + "]"))
				} else {
					b.WriteString(DimStyle.Render(" " + label + " "))
				}
			}
		}
		b.WriteString("\n")
	}

	b.WriteString("\n" + DimStyle.Render("  j/k navigate, h/l or arrows adjust, esc back"))

	return b.String()
}
