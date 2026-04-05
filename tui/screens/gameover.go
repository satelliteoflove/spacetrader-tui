package screens

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/the4ofus/spacetrader-tui/internal/economy"
	"github.com/the4ofus/spacetrader-tui/internal/game"
)

type GameOverScreen struct {
	gs *game.GameState
}

func NewGameOverScreen(gs *game.GameState) *GameOverScreen {
	return &GameOverScreen{gs: gs}
}

func (s *GameOverScreen) Init() tea.Cmd { return nil }

func (s *GameOverScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if key.Matches(msg, Keys.Enter) || key.Matches(msg, Keys.Back) {
			return s, tea.Quit
		}
	}
	return s, nil
}

func (s *GameOverScreen) View() string {
	var b strings.Builder

	if s.gs.EndStatus == game.StatusRetired {
		b.WriteString(SuccessStyle.Render("\n  CONGRATULATIONS!") + "\n\n")
		b.WriteString(fmt.Sprintf("  Commander %s has retired to their own moon!\n\n", s.gs.Player.Name))

		score := economy.CalculateScore(s.gs)
		b.WriteString(fmt.Sprintf("  Days played:      %d\n", s.gs.Day))
		b.WriteString(fmt.Sprintf("  Final credits:    %d\n", s.gs.Player.Credits))
		b.WriteString(fmt.Sprintf("  Difficulty:       %s (%d%%)\n\n", s.gs.Difficulty, score.DiffPercent))

		b.WriteString(fmt.Sprintf("  Worth points:     %d\n", score.WorthPoints))
		b.WriteString(fmt.Sprintf("  Days penalty:     -%d\n", score.DaysPenalty))
		b.WriteString(fmt.Sprintf("  Difficulty mult:  x%d%%\n", score.DiffPercent))
		b.WriteString("  " + strings.Repeat("-", 24) + "\n")
		b.WriteString(SelectedStyle.Render(fmt.Sprintf("  FINAL SCORE:      %d", score.FinalScore)) + "\n")

		rating := scoreRating(score.FinalScore)
		b.WriteString(fmt.Sprintf("\n  Rating: %s\n", rating))
	} else {
		b.WriteString(DangerStyle.Render("\n  GAME OVER") + "\n\n")
		b.WriteString(fmt.Sprintf("  Commander %s's ship was destroyed on day %d.\n",
			s.gs.Player.Name, s.gs.Day))
	}

	b.WriteString("\n" + DimStyle.Render("  press enter to exit"))
	return b.String()
}

func scoreRating(score int) string {
	switch {
	case score >= 300:
		return "Elite Trader"
	case score >= 200:
		return "Master Trader"
	case score >= 150:
		return "Expert Trader"
	case score >= 100:
		return "Competent Trader"
	case score >= 50:
		return "Average Trader"
	default:
		return "Beginner Trader"
	}
}
