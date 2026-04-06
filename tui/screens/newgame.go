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
	stageDifficulty
	stageSkills
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
			s.stage = stageDifficulty
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
				return s, func() tea.Msg {
					return StartGameMsg{
						Name:       s.name,
						Skills:     s.skills,
						Difficulty: s.difficulty,
					}
				}
			}
		case key.Matches(msg, Keys.Back):
			s.stage = stageDifficulty
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
			s.skills = [formula.NumSkills]int{1, 1, 1, 1}
			s.remaining = formula.SkillPointsForDifficulty(s.difficulty) - 4
			s.skillIdx = 0
			s.stage = stageSkills
		case key.Matches(msg, Keys.Back):
			s.stage = stageNameInput
			s.nameInput.Focus()
			return s, textinput.Blink
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

	case stageDifficulty:
		b.WriteString(fmt.Sprintf("Commander %s - Select difficulty:\n\n", s.name))
		for i, name := range s.diffNames {
			if i == s.diffIdx {
				b.WriteString(fmt.Sprintf("  %s\n", SelectedStyle.Render("> "+name)))
			} else {
				b.WriteString(fmt.Sprintf("    %s\n", NormalStyle.Render(name)))
			}
		}

		b.WriteString("\n")
		b.WriteString(difficultyDescription(gamedata.Difficulty(s.diffIdx)))

		b.WriteString("\n" + DimStyle.Render("j/k to select, enter to confirm, esc back"))

	case stageSkills:
		b.WriteString(fmt.Sprintf("Commander %s [%s] - Allocate skill points (%d remaining)\n\n",
			s.name, s.difficulty, s.remaining))

		for i, name := range s.skillNames {
			bar := strings.Repeat("|", s.skills[i]) + strings.Repeat(".", formula.SkillMax-s.skills[i])
			line := fmt.Sprintf("  %-10s [%s] %d", name, bar, s.skills[i])
			if i == s.skillIdx {
				b.WriteString(SelectedStyle.Render("> "+line) + "\n")
			} else {
				b.WriteString(NormalStyle.Render("  "+line) + "\n")
			}
		}

		b.WriteString("\n")
		b.WriteString(CyanStyle.Render("  "+s.skillNames[s.skillIdx]) + "\n")
		b.WriteString("  " + skillDescription(s.skillIdx, s.skills[s.skillIdx]) + "\n")

		b.WriteString("\n" + DimStyle.Render("j/k to select, h/l to adjust, enter when done, esc back"))
	}

	return b.String()
}

func skillDescription(idx, level int) string {
	switch idx {
	case formula.SkillPilot:
		chance := 30 + level*5
		return fmt.Sprintf("Flee from encounters. Current: %d%% escape chance.", chance)
	case formula.SkillFighter:
		return fmt.Sprintf("Combat damage and accuracy. Level %d.", level)
	case formula.SkillTrader:
		discount := level
		if discount > 10 {
			discount = 10
		}
		chance := 20 + level*5
		return fmt.Sprintf("Price discount: %d%%. Negotiate: %d%% success.", discount, chance)
	case formula.SkillEngineer:
		repair := level / 2
		if repair < 1 {
			repair = 1
		}
		return fmt.Sprintf("Auto-repairs %d hull per warp.", repair)
	}
	return ""
}

func difficultyDescription(d gamedata.Difficulty) string {
	points := formula.SkillPointsForDifficulty(d)
	var pirates, scoreMult string
	switch d {
	case gamedata.DiffBeginner:
		pirates = "Very weak"
		scoreMult = "x0.5"
	case gamedata.DiffEasy:
		pirates = "Weak"
		scoreMult = "x0.75"
	case gamedata.DiffNormal:
		pirates = "Average"
		scoreMult = "x1.0"
	case gamedata.DiffHard:
		pirates = "Strong"
		scoreMult = "x1.3"
	case gamedata.DiffImpossible:
		pirates = "Very strong"
		scoreMult = "x1.6"
	}
	return fmt.Sprintf("  %s  %d skill points\n  %s  Pirates: %s\n  %s  Score: %s\n",
		CyanStyle.Render("Skills:"), points,
		CyanStyle.Render("Danger:"), pirates,
		CyanStyle.Render("Score: "), scoreMult)
}
