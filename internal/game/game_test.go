package game_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/the4ofus/spacetrader-tui/internal/data"
	"github.com/the4ofus/spacetrader-tui/internal/formula"
	"github.com/the4ofus/spacetrader-tui/internal/game"
	"github.com/the4ofus/spacetrader-tui/internal/gamedata"
)

func loadData(t *testing.T) *gamedata.GameData {
	t.Helper()
	gd, err := data.LoadAll(os.DirFS("../../data"))
	if err != nil {
		t.Fatalf("LoadAll: %v", err)
	}
	return gd
}

func newTestGame(t *testing.T) *game.GameState {
	t.Helper()
	gd := loadData(t)
	skills := [formula.NumSkills]int{4, 4, 4, 4}
	return game.NewGameWithSeed(gd, "TestPlayer", skills, gamedata.DiffNormal, 42)
}

func TestNewGameDefaults(t *testing.T) {
	gs := newTestGame(t)

	if gs.Player.Name != "TestPlayer" {
		t.Errorf("name: got %q, want TestPlayer", gs.Player.Name)
	}
	if gs.Player.Credits != formula.StartingCredits {
		t.Errorf("credits: got %d, want %d", gs.Player.Credits, formula.StartingCredits)
	}
	if gs.Day != 1 {
		t.Errorf("day: got %d, want 1", gs.Day)
	}
	if gs.EndStatus != game.StatusPlaying {
		t.Errorf("status: got %d, want StatusPlaying", gs.EndStatus)
	}
	if gs.Player.Ship.TypeID != int(formula.StartingShip) {
		t.Errorf("ship type: got %d, want %d", gs.Player.Ship.TypeID, formula.StartingShip)
	}
}

func TestNewGameStartingSystem(t *testing.T) {
	gs := newTestGame(t)

	sys := gs.CurrentSystem()
	if sys.TechLevel < gamedata.TechEarlyIndustrial {
		t.Errorf("starting system %q has tech %v, expected >= Early Industrial", sys.Name, sys.TechLevel)
	}
	if sys.PoliticalSystem == gamedata.PolAnarchy {
		t.Errorf("starting system %q is Anarchy", sys.Name)
	}
	if !gs.Systems[gs.CurrentSystemID].Visited {
		t.Error("starting system not marked as visited")
	}
}

func TestNewGameMarkets(t *testing.T) {
	gs := newTestGame(t)

	marketsWithGoods := 0
	for i := range gs.Systems {
		hasAny := false
		for g := 0; g < game.NumGoods; g++ {
			if gs.Systems[i].Prices[g] > 0 {
				hasAny = true
			}
		}
		if hasAny {
			marketsWithGoods++
		}
	}
	if marketsWithGoods == 0 {
		t.Error("no systems have any goods priced")
	}
}

func TestNewGameStartingWeapon(t *testing.T) {
	gs := newTestGame(t)

	if len(gs.Player.Ship.Weapons) != 1 {
		t.Fatalf("expected 1 starting weapon, got %d", len(gs.Player.Ship.Weapons))
	}
	if gs.Player.Ship.Weapons[0] != 0 {
		t.Errorf("starting weapon should be Pulse Laser (id 0), got %d", gs.Player.Ship.Weapons[0])
	}
}

func TestNewGameDeterministic(t *testing.T) {
	gd := loadData(t)
	skills := [formula.NumSkills]int{4, 4, 4, 4}

	gs1 := game.NewGameWithSeed(gd, "Test", skills, gamedata.DiffNormal, 12345)
	gs2 := game.NewGameWithSeed(gd, "Test", skills, gamedata.DiffNormal, 12345)

	if gs1.CurrentSystemID != gs2.CurrentSystemID {
		t.Error("same seed should produce same starting system")
	}
	for i := range gs1.Systems {
		for g := 0; g < game.NumGoods; g++ {
			if gs1.Systems[i].Prices[g] != gs2.Systems[i].Prices[g] {
				t.Errorf("system %d good %d: prices differ with same seed", i, g)
			}
		}
	}
}

func TestSaveLoadRoundTrip(t *testing.T) {
	gs := newTestGame(t)

	gs.Player.Credits = 5000
	gs.Player.Cargo[0] = 3
	gs.Player.Cargo[4] = 7
	gs.Player.PoliceRecord = -15
	gs.Player.Reputation = 5
	gs.Day = 42

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test_save.json")

	if err := game.Save(gs, path); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded, err := game.Load(path, gs.Data)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if loaded.Player.Name != gs.Player.Name {
		t.Errorf("name: got %q, want %q", loaded.Player.Name, gs.Player.Name)
	}
	if loaded.Player.Credits != 5000 {
		t.Errorf("credits: got %d, want 5000", loaded.Player.Credits)
	}
	if loaded.Player.Cargo[0] != 3 {
		t.Errorf("cargo[0]: got %d, want 3", loaded.Player.Cargo[0])
	}
	if loaded.Player.Cargo[4] != 7 {
		t.Errorf("cargo[4]: got %d, want 7", loaded.Player.Cargo[4])
	}
	if loaded.Day != 42 {
		t.Errorf("day: got %d, want 42", loaded.Day)
	}
	if loaded.CurrentSystemID != gs.CurrentSystemID {
		t.Errorf("current system: got %d, want %d", loaded.CurrentSystemID, gs.CurrentSystemID)
	}
	if loaded.Player.PoliceRecord != -15 {
		t.Errorf("police record: got %d, want -15", loaded.Player.PoliceRecord)
	}

	for i := range loaded.Systems {
		for g := 0; g < game.NumGoods; g++ {
			if loaded.Systems[i].Prices[g] != gs.Systems[i].Prices[g] {
				t.Errorf("system %d good %d price mismatch after load", i, g)
			}
		}
	}
}

func TestPlayerCargo(t *testing.T) {
	gs := newTestGame(t)

	gs.Player.Cargo[0] = 5
	gs.Player.Cargo[3] = 3

	if got := gs.Player.TotalCargo(); got != 8 {
		t.Errorf("TotalCargo: got %d, want 8", got)
	}

	dp := &game.GameDataProvider{Data: gs.Data}
	cap := gs.Player.CargoCapacity(dp)
	if cap != 15 {
		t.Errorf("CargoCapacity: got %d, want 15 (Gnat)", cap)
	}

	free := gs.Player.FreeCargo(dp)
	if free != 7 {
		t.Errorf("FreeCargo: got %d, want 7", free)
	}
}
