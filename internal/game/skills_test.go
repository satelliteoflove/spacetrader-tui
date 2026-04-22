package game_test

import (
	"os"
	"testing"

	"github.com/the4ofus/spacetrader-tui/internal/data"
	"github.com/the4ofus/spacetrader-tui/internal/formula"
	"github.com/the4ofus/spacetrader-tui/internal/game"
	"github.com/the4ofus/spacetrader-tui/internal/gamedata"
)

func newSkillsTestGame(t *testing.T) *game.GameState {
	t.Helper()
	gd, err := data.LoadAll(os.DirFS("../../data"))
	if err != nil {
		t.Fatalf("LoadAll: %v", err)
	}
	skills := [formula.NumSkills]int{4, 4, 4, 4}
	return game.NewGameWithSeed(gd, "Test", skills, gamedata.DiffNormal, 42)
}

func findEquipIDByName(gs *game.GameState, name string) int {
	for i, eq := range gs.Data.Equipment {
		if eq.Name == name {
			return i
		}
	}
	return -1
}

func TestHasTradeAnalyzerFalseByDefault(t *testing.T) {
	gs := newSkillsTestGame(t)
	gs.Player.Ship.Gadgets = nil
	if game.HasTradeAnalyzer(gs) {
		t.Error("expected no trade analyzer on fresh ship")
	}
}

func TestHasTradeAnalyzerTrueWhenInstalled(t *testing.T) {
	gs := newSkillsTestGame(t)
	id := findEquipIDByName(gs, game.TradeAnalyzerName)
	if id < 0 {
		t.Fatal("Trade Analyzer not found in loaded equipment")
	}
	gs.Player.Ship.Gadgets = []int{id}
	if !game.HasTradeAnalyzer(gs) {
		t.Error("expected HasTradeAnalyzer to be true when gadget equipped")
	}
}

func TestHasTradeAnalyzerIgnoresOtherGadgets(t *testing.T) {
	gs := newSkillsTestGame(t)
	cargoID := findEquipIDByName(gs, "Extra Cargo Bays")
	if cargoID < 0 {
		t.Fatal("Extra Cargo Bays not found")
	}
	gs.Player.Ship.Gadgets = []int{cargoID}
	if game.HasTradeAnalyzer(gs) {
		t.Error("Extra Cargo Bays should not count as Trade Analyzer")
	}
}
