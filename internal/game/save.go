package game

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"

	"github.com/the4ofus/spacetrader-tui/internal/galaxy"
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

func AutosavePath() (string, error) {
	dir, err := DefaultSaveDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "autosave.json"), nil
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

func Autosave(gs *GameState) error {
	path, err := AutosavePath()
	if err != nil {
		return err
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(gs, "", "  ")
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, b, 0644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
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

	if gs.SaveVersion < CurrentSaveVersion {
		return nil, fmt.Errorf("incompatible save (version %d, need %d)", gs.SaveVersion, CurrentSaveVersion)
	}

	gs.Rand = rand.New(rand.NewSource(gs.Seed))
	data.Systems = galaxy.Generate(gs.Seed)
	gs.Data = data

	numSystems := len(data.Systems)
	for len(gs.Systems) < numSystems {
		idx := len(gs.Systems)
		gs.Systems = append(gs.Systems, SystemState{})
		RefreshSystemPrices(gs, idx)
	}
	if len(gs.Systems) > numSystems {
		gs.Systems = gs.Systems[:numSystems]
	}

	validateSave(gs)

	return gs, nil
}

func validateSave(gs *GameState) {
	numSys := len(gs.Data.Systems)
	numEquip := len(gs.Data.Equipment)

	if gs.CurrentSystemID < 0 || gs.CurrentSystemID >= numSys {
		gs.CurrentSystemID = 0
	}

	gs.Player.Ship.Weapons = filterValidIDs(gs.Player.Ship.Weapons, numEquip)
	gs.Player.Ship.Shields = filterValidIDs(gs.Player.Ship.Shields, numEquip)
	gs.Player.Ship.Gadgets = filterValidIDs(gs.Player.Ship.Gadgets, numEquip)

	if gs.Player.Ship.TypeID < 0 || gs.Player.Ship.TypeID >= len(gs.Data.Ships) {
		gs.Player.Ship.TypeID = 0
	}

	valid := gs.Wormholes[:0]
	for _, wh := range gs.Wormholes {
		if wh.SystemA >= 0 && wh.SystemA < numSys && wh.SystemB >= 0 && wh.SystemB < numSys {
			valid = append(valid, wh)
		}
	}
	gs.Wormholes = valid

	validBM := gs.Bookmarks[:0]
	for _, bm := range gs.Bookmarks {
		if bm.SystemIdx >= 0 && bm.SystemIdx < numSys {
			validBM = append(validBM, bm)
		}
	}
	gs.Bookmarks = validBM

	validNews := gs.NewsLog[:0]
	for _, ne := range gs.NewsLog {
		if ne.SystemIdx >= 0 && ne.SystemIdx < numSys {
			validNews = append(validNews, ne)
		}
	}
	gs.NewsLog = validNews
}

func filterValidIDs(ids []int, maxLen int) []int {
	valid := ids[:0]
	for _, id := range ids {
		if id >= 0 && id < maxLen {
			valid = append(valid, id)
		}
	}
	return valid
}
