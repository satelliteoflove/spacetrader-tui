package game

import (
	"encoding/json"
	"math/rand"
	"os"
	"path/filepath"

	"github.com/the4ofus/spacetrader-tui/internal/gamedata"
)

func DefaultSaveDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".spacetrader")
	return dir, os.MkdirAll(dir, 0755)
}

func DefaultSavePath() (string, error) {
	dir, err := DefaultSaveDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "save.json"), nil
}

func Save(gs *GameState, path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(gs, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0644)
}

func Load(path string, data *gamedata.GameData) (*GameState, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	gs := &GameState{}
	if err := json.Unmarshal(b, gs); err != nil {
		return nil, err
	}
	gs.Rand = rand.New(rand.NewSource(gs.Seed))
	gs.Data = data
	return gs, nil
}
