package game

import (
	"testing"

	"github.com/the4ofus/spacetrader-tui/internal/formula"
	"github.com/the4ofus/spacetrader-tui/internal/gamedata"
)

func TestFehlerFabricRip(t *testing.T) {
	gs := newTestGameState(t)
	gs.Day = 45

	gs.SetQuestState(QuestFehler, QuestAvailable)
	gs.SetQuestProgress(QuestFehler, gs.Day)

	gs.Day += 6

	otherSys := 0
	deneb := findSystem(gs, "Deneb")
	if otherSys == deneb {
		otherSys = 1
	}
	gs.CurrentSystemID = otherSys
	events := CheckQuestsOnArrival(gs)

	if gs.Quests.FabricRipDays != 25 {
		t.Errorf("expected FabricRipDays 25, got %d", gs.Quests.FabricRipDays)
	}
	if gs.QuestState(QuestFabricRip) != QuestActive {
		t.Errorf("expected QuestFabricRip Active, got %d", gs.QuestState(QuestFabricRip))
	}

	found := false
	for _, e := range events {
		if e.Title == "Experiment Failed!" {
			found = true
		}
	}
	if !found {
		t.Error("expected Experiment Failed! event")
	}
}

func TestReactorPickup(t *testing.T) {
	gs := newTestGameState(t)
	gs.Day = 50
	gs.Player.PoliceRecord = -10
	gs.Player.Reputation = 50

	gs.SetQuestState(QuestReactor, QuestAvailable)

	dp := &GameDataProvider{Data: gs.Data}
	freeBefore := gs.Player.FreeCargo(dp)
	if freeBefore < ReactorTotalBays {
		gs.Player.Cargo = [10]int{}
	}

	result := resolveQuestChainAction(gs, "Reactor Delivery", 0)
	if result == "" {
		t.Fatal("expected non-empty result from reactor acceptance")
	}

	if gs.QuestState(QuestReactor) != QuestActive {
		t.Errorf("expected Reactor quest Active, got %d", gs.QuestState(QuestReactor))
	}
	if gs.QuestProgress(QuestReactor) != ReactorStatusFuelOk {
		t.Errorf("expected reactor status %d, got %d", ReactorStatusFuelOk, gs.QuestProgress(QuestReactor))
	}
	if !ReactorOnBoard(gs) {
		t.Error("expected ReactorOnBoard to be true")
	}
	bays := ReactorCargoBays(gs)
	if bays != ReactorTotalBays {
		t.Errorf("expected %d cargo bays consumed, got %d", ReactorTotalBays, bays)
	}
}

func TestReactorCargoBays(t *testing.T) {
	gs := newTestGameState(t)
	gs.Player.Cargo = [10]int{}

	gs.SetQuestState(QuestReactor, QuestActive)

	tests := []struct {
		status      int
		expectedBays int
	}{
		{1, 15},
		{2, 15},
		{3, 14},
		{5, 13},
		{10, 11},
		{15, 8},
		{19, 6},
	}

	for _, tc := range tests {
		gs.SetQuestProgress(QuestReactor, tc.status)
		bays := ReactorCargoBays(gs)
		if bays != tc.expectedBays {
			t.Errorf("at status %d: expected %d bays, got %d", tc.status, tc.expectedBays, bays)
		}
	}

	gs.SetQuestProgress(QuestReactor, ReactorStatusDate)
	if ReactorOnBoard(gs) {
		t.Error("reactor should not be 'on board' at meltdown status")
	}
	if ReactorCargoBays(gs) != 0 {
		t.Error("cargo bays should be 0 when reactor is not on board")
	}
}

func TestReactorDeliveryToNix(t *testing.T) {
	gs := newTestGameState(t)
	gs.Player.Cargo = [10]int{}

	gs.SetQuestState(QuestReactor, QuestActive)
	gs.SetQuestProgress(QuestReactor, ReactorStatusFuelOk)

	gs.Player.Ship.Weapons = []int{}

	nix := findSystem(gs, "Nix")
	if nix < 0 {
		t.Fatal("Nix not found")
	}
	gs.CurrentSystemID = nix
	events := CheckQuestsOnArrival(gs)

	delivered := false
	for _, e := range events {
		if e.Title == "Reactor Delivered!" {
			delivered = true
		}
	}
	if !delivered {
		t.Error("expected Reactor Delivered! event")
	}

	shipDef := gs.Data.Ships[gs.Player.Ship.TypeID]
	if shipDef.WeaponSlots > 0 {
		if gs.QuestProgress(QuestReactor) != ReactorStatusDone {
			t.Errorf("with free weapon slot, expected status Done (%d), got %d", ReactorStatusDone, gs.QuestProgress(QuestReactor))
		}
		if gs.QuestState(QuestReactor) != QuestComplete {
			t.Error("expected reactor quest Complete after laser installed")
		}
		morgansIdx := findEquipByName(gs, "Morgan's Laser")
		hasLaser := false
		for _, w := range gs.Player.Ship.Weapons {
			if w == morgansIdx {
				hasLaser = true
			}
		}
		if !hasLaser {
			t.Error("expected Morgan's Laser to be installed")
		}
	}

	if ReactorOnBoard(gs) {
		t.Error("reactor should no longer be on board after delivery")
	}
}

func TestReactorDeliveryNoWeaponSlot(t *testing.T) {
	gs := newTestGameState(t)
	gs.Player.Cargo = [10]int{}

	gs.SetQuestState(QuestReactor, QuestActive)
	gs.SetQuestProgress(QuestReactor, ReactorStatusFuelOk)

	shipDef := gs.Data.Ships[gs.Player.Ship.TypeID]
	gs.Player.Ship.Weapons = nil
	for i := 0; i < shipDef.WeaponSlots; i++ {
		gs.Player.Ship.Weapons = append(gs.Player.Ship.Weapons, 0)
	}

	nix := findSystem(gs, "Nix")
	if nix < 0 {
		t.Fatal("Nix not found")
	}
	gs.CurrentSystemID = nix
	CheckQuestsOnArrival(gs)

	if gs.QuestProgress(QuestReactor) != ReactorStatusDelivered {
		t.Errorf("with full weapon slots, expected status Delivered (%d), got %d", ReactorStatusDelivered, gs.QuestProgress(QuestReactor))
	}
	if len(gs.Quests.PendingRewards) == 0 {
		t.Error("expected pending reward for Morgan's Laser")
	}
}

func TestReactorMeltdownGameOver(t *testing.T) {
	gs := newTestGameState(t)
	gs.Player.HasEscapePod = false

	gs.SetQuestState(QuestReactor, QuestActive)
	gs.SetQuestProgress(QuestReactor, ReactorStatusDate)

	otherSys := 0
	nix := findSystem(gs, "Nix")
	if otherSys == nix {
		otherSys = 1
	}
	gs.CurrentSystemID = otherSys
	events := CheckQuestsOnArrival(gs)

	if gs.EndStatus != StatusDead {
		t.Errorf("expected StatusDead, got %d", gs.EndStatus)
	}

	found := false
	for _, e := range events {
		if e.Title == "Reactor Meltdown!" {
			found = true
		}
	}
	if !found {
		t.Error("expected Reactor Meltdown! event")
	}
}

func TestReactorMeltdownEscapePod(t *testing.T) {
	gs := newTestGameState(t)
	gs.Player.HasEscapePod = true

	gs.SetQuestState(QuestReactor, QuestActive)
	gs.SetQuestProgress(QuestReactor, ReactorStatusDate)

	otherSys := 0
	nix := findSystem(gs, "Nix")
	if otherSys == nix {
		otherSys = 1
	}
	gs.CurrentSystemID = otherSys
	events := CheckQuestsOnArrival(gs)

	if gs.EndStatus != StatusPlaying {
		t.Errorf("expected StatusPlaying (survived via pod), got %d", gs.EndStatus)
	}
	if gs.Player.Ship.TypeID != ShipFlea {
		t.Errorf("expected Flea ship, got type %d", gs.Player.Ship.TypeID)
	}

	totalCargo := 0
	for _, qty := range gs.Player.Cargo {
		totalCargo += qty
	}
	if totalCargo != 0 {
		t.Errorf("expected cleared cargo, got total %d", totalCargo)
	}

	found := false
	for _, e := range events {
		if e.Title == "Reactor Meltdown!" {
			found = true
		}
	}
	if !found {
		t.Error("expected Reactor Meltdown! event")
	}
}

func TestReactorWildMutualExclusion(t *testing.T) {
	gs := newTestGameState(t)
	gs.Player.Cargo = [10]int{}

	gs.SetQuestState(QuestWild, QuestActive)
	gs.Player.Crew = append(gs.Player.Crew, Mercenary{
		Name: "Wild", Skills: [formula.NumSkills]int{7, 10, 2, 5}, SystemIdx: -1, IsQuest: true,
	})
	gs.SetQuestState(QuestReactor, QuestAvailable)

	result := resolveQuestChainAction(gs, "Reactor Delivery", 0)
	if result == "" {
		t.Fatal("expected non-empty result")
	}

	if gs.QuestState(QuestReactor) != QuestActive {
		t.Errorf("expected Reactor quest Active, got %d", gs.QuestState(QuestReactor))
	}
	if gs.QuestState(QuestWild) != QuestUnavailable {
		t.Errorf("expected Wild quest reset to Unavailable, got %d", gs.QuestState(QuestWild))
	}
	if HasQuestCrew(gs, "Wild") {
		t.Error("expected Wild removed from crew")
	}
}

func TestJaporiCompletion(t *testing.T) {
	gs := newTestGameState(t)

	gs.SetQuestState(QuestJapori, QuestActive)
	gs.Player.Skills = [4]int{3, 3, 3, 3}

	skillsBefore := 0
	for _, s := range gs.Player.Skills {
		skillsBefore += s
	}

	medIdx := int(gamedata.GoodMedicine)
	gs.Player.Cargo[medIdx] = 10

	japori := findSystem(gs, "Japori")
	if japori < 0 {
		t.Fatal("Japori not found")
	}
	gs.CurrentSystemID = japori
	events := CheckQuestsOnArrival(gs)

	if gs.QuestState(QuestJapori) != QuestComplete {
		t.Error("expected Japori quest Complete")
	}
	if gs.Player.Cargo[medIdx] != 0 {
		t.Errorf("expected medicine removed, got %d", gs.Player.Cargo[medIdx])
	}

	skillsAfter := 0
	for _, s := range gs.Player.Skills {
		skillsAfter += s
	}
	if skillsAfter <= skillsBefore {
		t.Errorf("expected skills to increase: before=%d, after=%d", skillsBefore, skillsAfter)
	}

	found := false
	for _, e := range events {
		if e.Title == "Japori Disease - Complete!" {
			found = true
		}
	}
	if !found {
		t.Error("expected Japori Disease - Complete! event")
	}
}

func TestJaporiNoMedicineFeedback(t *testing.T) {
	gs := newTestGameState(t)

	gs.SetQuestState(QuestJapori, QuestActive)

	medIdx := int(gamedata.GoodMedicine)
	gs.Player.Cargo[medIdx] = 3

	japori := findSystem(gs, "Japori")
	if japori < 0 {
		t.Fatal("Japori not found")
	}
	gs.CurrentSystemID = japori
	events := CheckQuestsOnArrival(gs)

	if gs.QuestState(QuestJapori) == QuestComplete {
		t.Error("Japori quest should not be complete with < 10 medicine")
	}

	found := false
	for _, e := range events {
		if e.Title == "Japori Disease" {
			found = true
		}
	}
	if !found {
		t.Error("expected Japori Disease feedback event when arriving without enough medicine")
	}
}

func TestMoonForSaleVisibility(t *testing.T) {
	gs := newTestGameState(t)

	gs.Player.Credits = 100
	gs.Player.LoanBalance = 0
	gs.SetQuestState(QuestMoonForSale, QuestUnavailable)
	CheckQuestsOnArrival(gs)
	if gs.QuestState(QuestMoonForSale) != QuestUnavailable {
		t.Error("moon should stay Unavailable with low net worth")
	}

	gs.Player.Credits = 500000
	CheckQuestsOnArrival(gs)
	if gs.QuestState(QuestMoonForSale) != QuestAvailable {
		t.Error("moon should become Available with high net worth")
	}
}
