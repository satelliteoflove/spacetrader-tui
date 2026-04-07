package main

import (
	"embed"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/the4ofus/spacetrader-tui/internal/data"
	"github.com/the4ofus/spacetrader-tui/internal/game"
	"github.com/the4ofus/spacetrader-tui/tui"
	"github.com/the4ofus/spacetrader-tui/tui/screens"
)

//go:embed data/*.json
var dataFS embed.FS

func main() {
	gd, err := data.LoadAllFromEmbed(dataFS)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load game data: %v\n", err)
		os.Exit(1)
	}

	cfg := game.LoadConfig()
	screens.InitStyles(cfg.ColorblindMode)
	tui.InitStatusStyles(cfg.ColorblindMode)
	screens.ApplyAnimationSettings(cfg)

	m := tui.NewModel(gd, cfg)
	p := tea.NewProgram(m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
