package encounter_test

import (
	"os"
	"testing"

	"github.com/the4ofus/spacetrader-tui/internal/data"
	"github.com/the4ofus/spacetrader-tui/internal/encounter"
	"github.com/the4ofus/spacetrader-tui/internal/formula"
	"github.com/the4ofus/spacetrader-tui/internal/game"
	"github.com/the4ofus/spacetrader-tui/internal/gamedata"
)

func newTestGame(t *testing.T, seed int64) *game.GameState {
	t.Helper()
	gd, err := data.LoadAll(os.DirFS("../../data"))
	if err != nil {
		t.Fatalf("LoadAll: %v", err)
	}
	skills := [formula.NumSkills]int{4, 4, 4, 4}
	return game.NewGameWithSeed(gd, "Test", skills, gamedata.DiffNormal, seed)
}

func TestGenerateReturnsEncounters(t *testing.T) {
	encounters := map[encounter.EncounterType]int{}

	for seed := int64(0); seed < 500; seed++ {
		gs := newTestGame(t, seed)
		enc := encounter.Generate(gs)
		if enc != nil {
			encounters[enc.Type]++
		}
	}

	if encounters[encounter.EncPolice] == 0 {
		t.Error("no police encounters generated in 500 seeds")
	}
	if encounters[encounter.EncPirate] == 0 {
		t.Error("no pirate encounters generated in 500 seeds")
	}
	if encounters[encounter.EncTrader] == 0 {
		t.Error("no trader encounters generated in 500 seeds")
	}
}

func TestPoliceComplyClean(t *testing.T) {
	gs := newTestGame(t, 42)
	enc := encounter.NewPoliceEncounter()

	outcome := encounter.Resolve(gs, enc, encounter.ActionComply)
	if outcome.RecordChange < 0 {
		t.Error("clean player complying should not worsen record")
	}
}

func TestPoliceComplyWithIllegal(t *testing.T) {
	gs := newTestGame(t, 42)
	gs.Player.Cargo[int(gamedata.GoodFirearms)] = 3

	enc := encounter.NewPoliceEncounter()
	outcome := encounter.Resolve(gs, enc, encounter.ActionComply)

	if outcome.RecordChange >= 0 {
		t.Error("carrying illegal goods and complying should worsen record")
	}
	if gs.Player.Cargo[int(gamedata.GoodFirearms)] != 0 {
		t.Error("illegal goods should be confiscated")
	}
	if outcome.CreditsChange >= 0 {
		t.Error("should be fined")
	}
}

func TestPirateFight(t *testing.T) {
	gs := newTestGame(t, 42)
	gs.Player.Skills[formula.SkillFighter] = 10

	enc := encounter.NewPirateEncounter()
	startCredits := gs.Player.Credits

	encounter.Resolve(gs, enc, encounter.ActionFight)

	creditsChanged := gs.Player.Credits != startCredits
	if !creditsChanged {
		t.Error("credits should change after pirate fight")
	}
}

func TestPirateSurrender(t *testing.T) {
	gs := newTestGame(t, 42)
	gs.Player.Cargo[0] = 5
	gs.Player.Cargo[1] = 5

	enc := encounter.NewPirateEncounter()
	outcome := encounter.Resolve(gs, enc, encounter.ActionSurrender)

	if outcome.CreditsChange >= 0 {
		t.Error("surrender should lose credits")
	}
}

func TestTraderTrade(t *testing.T) {
	gs := newTestGame(t, 42)
	gs.Player.Credits = 10000

	enc := encounter.NewTraderEncounter()
	outcome := encounter.Resolve(gs, enc, encounter.ActionTrade)

	if outcome.Message == "" {
		t.Error("should have a message")
	}
}

func TestTraderDecline(t *testing.T) {
	gs := newTestGame(t, 42)

	enc := encounter.NewTraderEncounter()
	outcome := encounter.Resolve(gs, enc, encounter.ActionDecline)

	if outcome.Message != "Declined to trade." {
		t.Errorf("unexpected message: %q", outcome.Message)
	}
}

func TestEncounterActions(t *testing.T) {
	police := encounter.NewPoliceEncounter()
	if len(police.Actions) != 3 {
		t.Errorf("police should have 3 actions, got %d", len(police.Actions))
	}

	pirate := encounter.NewPirateEncounter()
	if len(pirate.Actions) != 3 {
		t.Errorf("pirate should have 3 actions, got %d", len(pirate.Actions))
	}

	trader := encounter.NewTraderEncounter()
	if len(trader.Actions) != 2 {
		t.Errorf("trader should have 2 actions, got %d", len(trader.Actions))
	}
}
