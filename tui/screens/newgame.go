package screens

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/the4ofus/spacetrader-tui/internal/formula"
	"github.com/the4ofus/spacetrader-tui/internal/gamedata"
)

type newGameStage int

const (
	stageNameInput newGameStage = iota
	stageSkills
	stageDifficulty
)

type NewGameScreen struct {
	stage      newGameStage
	nameInput  textinput.Model
	name       string
	skills     [formula.NumSkills]int
	skillNames [formula.NumSkills]string
	skillIdx   int
	remaining  int
	difficulty gamedata.Difficulty
	diffIdx    int
	diffNames  []string
}

func NewNewGameScreen() *NewGameScreen {
	ti := textinput.New()
	ti.Placeholder = "Commander"
	ti.Focus()
	ti.CharLimit = 20

	return &NewGameScreen{
		stage:      stageNameInput,
		nameInput:  ti,
		skillNames: [formula.NumSkills]string{"Pilot", "Fighter", "Trader", "Engineer"},
		skills:     [formula.NumSkills]int{1, 1, 1, 1},
		remaining:  12,
		difficulty: gamedata.DiffNormal,
		diffIdx:    2,
		diffNames:  []string{"Beginner", "Easy", "Normal", "Hard", "Impossible"},
	}
}

func (s *NewGameScreen) Init() tea.Cmd {
	return textinput.Blink
}

func (s *NewGameScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch s.stage {
	case stageNameInput:
		return s.updateName(msg)
	case stageSkills:
		return s.updateSkills(msg)
	case stageDifficulty:
		return s.updateDifficulty(msg)
	}
	return s, nil
}

func (s *NewGameScreen) updateName(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if key.Matches(msg, Keys.Enter) {
			name := strings.TrimSpace(s.nameInput.Value())
			if name == "" {
				name = "Commander"
			}
			s.name = name
			s.stage = stageSkills
			s.remaining = formula.SkillPointsForDifficulty(s.difficulty) - 4
			return s, nil
		}
		if key.Matches(msg, Keys.Back) {
			return s, func() tea.Msg { return NavigateMsg{Screen: ScreenTitle} }
		}
	}
	var cmd tea.Cmd
	s.nameInput, cmd = s.nameInput.Update(msg)
	return s, cmd
}

func (s *NewGameScreen) updateSkills(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, Keys.Up):
			s.skillIdx = wrapCursor(s.skillIdx, -1, formula.NumSkills)
		case key.Matches(msg, Keys.Down):
			s.skillIdx = wrapCursor(s.skillIdx, 1, formula.NumSkills)
		case msg.String() == "right" || msg.String() == "l":
			if s.remaining > 0 && s.skills[s.skillIdx] < formula.SkillMax {
				s.skills[s.skillIdx]++
				s.remaining--
			}
		case msg.String() == "left" || msg.String() == "h":
			if s.skills[s.skillIdx] > formula.SkillMin {
				s.skills[s.skillIdx]--
				s.remaining++
			}
		case key.Matches(msg, Keys.Enter):
			if s.remaining == 0 {
				s.stage = stageDifficulty
			}
		case key.Matches(msg, Keys.Back):
			s.stage = stageNameInput
			s.nameInput.Focus()
			return s, textinput.Blink
		}
	}
	return s, nil
}

func (s *NewGameScreen) updateDifficulty(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, Keys.Up):
			s.diffIdx = wrapCursor(s.diffIdx, -1, len(s.diffNames))
		case key.Matches(msg, Keys.Down):
			s.diffIdx = wrapCursor(s.diffIdx, 1, len(s.diffNames))
		case key.Matches(msg, Keys.Enter):
			s.difficulty = gamedata.Difficulty(s.diffIdx)
			return s, func() tea.Msg {
				return StartGameMsg{
					Name:       s.name,
					Skills:     s.skills,
					Difficulty: s.difficulty,
				}
			}
		case key.Matches(msg, Keys.Back):
			s.stage = stageSkills
		}
	}
	return s, nil
}

func (s *NewGameScreen) View() string {
	var b strings.Builder

	b.WriteString(TitleStyle.Render("CHARACTER CREATION") + "\n\n")

	switch s.stage {
	case stageNameInput:
		b.WriteString("Enter your name:\n\n")
		b.WriteString(s.nameInput.View() + "\n\n")
		b.WriteString(DimStyle.Render("enter to continue, esc to go back"))

	case stageSkills:
		b.WriteString(fmt.Sprintf("Commander %s - Allocate skill points (%d remaining)\n\n",
			s.name, s.remaining))

		for i, name := range s.skillNames {
			bar := strings.Repeat("|", s.skills[i]) + strings.Repeat(".", formula.SkillMax-s.skills[i])
			line := fmt.Sprintf("  %-10s [%s] %d", name, bar, s.skills[i])
			if i == s.skillIdx {
				b.WriteString(SelectedStyle.Render("> "+line) + "\n")
			} else {
				b.WriteString(NormalStyle.Render("  "+line) + "\n")
			}
		}

		b.WriteString("\n" + DimStyle.Render("j/k to select, h/l to adjust, enter when done"))

	case stageDifficulty:
		b.WriteString("Select difficulty:\n\n")
		for i, name := range s.diffNames {
			if i == s.diffIdx {
				b.WriteString(fmt.Sprintf("  %s\n", SelectedStyle.Render("> "+name)))
			} else {
				b.WriteString(fmt.Sprintf("    %s\n", NormalStyle.Render(name)))
			}
		}
		b.WriteString("\n" + DimStyle.Render("j/k to select, enter to confirm"))
	}

	return b.String()
}
