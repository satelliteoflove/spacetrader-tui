package game

import (
	"os"
	"testing"

	"github.com/the4ofus/spacetrader-tui/internal/data"
	"github.com/the4ofus/spacetrader-tui/internal/formula"
	"github.com/the4ofus/spacetrader-tui/internal/gamedata"
)

func newTestGameState(t *testing.T) *GameState {
	t.Helper()
	gd, err := data.LoadAll(os.DirFS("../../data"))
	if err != nil {
		t.Fatalf("LoadAll: %v", err)
	}
	skills := [formula.NumSkills]int{5, 5, 5, 5}
	gs := NewGameWithSeed(gd, "Test", skills, gamedata.DiffNormal, 42)
	return gs
}

func TestDragonflyFullPath(t *testing.T) {
	gs := newTestGameState(t)
	gs.Day = 25

	gs.SetQuestState(QuestDragonfly, QuestAvailable)
	gs.SetQuestProgress(QuestDragonfly, 0)
	gs.Player.Ship.Weapons = []int{0}

	path := []string{"Arouan", "Halley", "Regulus", "Linnet"}
	for i, name := range path {
		sysIdx := findSystem(gs, name)
		if sysIdx < 0 {
			t.Fatalf("system %s not found", name)
		}
		gs.CurrentSystemID = sysIdx
		events := CheckQuestsOnArrival(gs)

		if i < len(path)-1 {
			if gs.QuestProgress(QuestDragonfly) != i+1 {
				t.Errorf("step %d: expected progress %d, got %d", i, i+1, gs.QuestProgress(QuestDragonfly))
			}
			if gs.QuestState(QuestDragonfly) != QuestActive {
				t.Errorf("step %d: expected Active state", i)
			}
			found := false
			for _, e := range events {
				if e.Title == "Dragonfly Spotted" {
					found = true
				}
			}
			if !found {
				t.Errorf("step %d: expected Dragonfly Spotted event", i)
			}
		} else {
			found := false
			for _, e := range events {
				if e.Title == "Dragonfly Cornered!" {
					found = true
				}
			}
			if !found {
				t.Errorf("final step: expected Dragonfly Cornered! event")
			}

			gs.Quests.DragonflyHull = 1
			resolveDragonflyCombat(gs)

			if gs.QuestState(QuestDragonfly) != QuestComplete {
				t.Errorf("expected Complete after combat, got %d", gs.QuestState(QuestDragonfly))
			}
		}
	}
}

func TestDragonflyEquipmentPending(t *testing.T) {
	gs := newTestGameState(t)
	gs.Day = 25

	gs.SetQuestState(QuestDragonfly, QuestAvailable)
	gs.SetQuestProgress(QuestDragonfly, 0)

	fireflyIdx := -1
	for i, s := range gs.Data.Ships {
		if s.ShieldSlots > 0 {
			fireflyIdx = i
			break
		}
	}
	if fireflyIdx < 0 {
		t.Fatal("no ship with shield slots found")
	}
	gs.Player.Ship.TypeID = fireflyIdx
	shipDef := gs.Data.Ships[fireflyIdx]
	gs.Player.Ship.Hull = shipDef.Hull
	gs.Player.Ship.Weapons = []int{0}

	gs.Player.Ship.Shields = nil
	for len(gs.Player.Ship.Shields) < shipDef.ShieldSlots {
		gs.Player.Ship.Shields = append(gs.Player.Ship.Shields, 0)
	}

	path := []string{"Arouan", "Halley", "Regulus", "Linnet"}
	for _, name := range path {
		sysIdx := findSystem(gs, name)
		if sysIdx < 0 {
			t.Fatalf("system %s not found", name)
		}
		gs.CurrentSystemID = sysIdx
		CheckQuestsOnArrival(gs)
	}

	gs.Quests.DragonflyHull = 1
	resolveDragonflyCombat(gs)

	if gs.QuestState(QuestDragonfly) != QuestComplete {
		t.Fatalf("expected Complete, got %d", gs.QuestState(QuestDragonfly))
	}
	if len(gs.Quests.PendingRewards) == 0 {
		t.Fatal("expected pending reward for Lightning Shield")
	}

	gs.Player.Ship.Shields = gs.Player.Ship.Shields[:len(gs.Player.Ship.Shields)-1]

	linnet := findSystem(gs, "Linnet")
	gs.CurrentSystemID = linnet
	events := CheckQuestsOnArrival(gs)

	installed := false
	for _, e := range events {
		if e.Title == "Equipment Installed!" {
			installed = true
		}
	}
	if !installed {
		t.Error("expected Equipment Installed! event after freeing shield slot")
	}
	if len(gs.Quests.PendingRewards) != 0 {
		t.Error("expected pending rewards to be cleared")
	}
}

func TestSpaceMonsterAttrition(t *testing.T) {
	gs := newTestGameState(t)
	gs.Day = 20

	gs.SetQuestState(QuestSpaceMonster, QuestAvailable)
	gs.Quests.MonsterHull = MonsterMaxHull

	gs.Player.Skills[formula.SkillFighter] = 1
	gs.Player.Skills[formula.SkillEngineer] = 1
	gs.Player.Ship.Weapons = []int{0}
	shipDef := gs.Data.Ships[gs.Player.Ship.TypeID]
	gs.Player.Ship.Hull = shipDef.Hull

	acamar := findSystem(gs, "Acamar")
	if acamar < 0 {
		t.Fatal("Acamar not found")
	}
	gs.CurrentSystemID = acamar

	events := CheckQuestsOnArrival(gs)
	found := false
	for _, e := range events {
		if e.Title == "Space Monster!" {
			found = true
		}
	}
	if !found {
		t.Fatal("expected Space Monster! event at Acamar")
	}

	resolveMonsterCombat(gs)

	if gs.QuestState(QuestSpaceMonster) == QuestComplete {
		t.Skip("monster died in one fight (unlikely with low skills), can't test attrition")
	}

	if gs.Quests.MonsterHull >= MonsterMaxHull {
		t.Error("expected monster hull to decrease after combat")
	}
	if gs.Quests.MonsterHull <= 0 {
		t.Error("monster hull should be positive if quest is not complete")
	}

	damagedHull := gs.Quests.MonsterHull
	regenned := damagedHull * 105 / 100
	if regenned > MonsterMaxHull {
		regenned = MonsterMaxHull
	}
	if regenned <= damagedHull {
		t.Error("regen formula should produce hull increase when damaged")
	}
}

func TestSpaceMonsterNoWeapons(t *testing.T) {
	gs := newTestGameState(t)
	gs.Quests.MonsterHull = MonsterMaxHull
	gs.Player.Ship.Weapons = []int{}

	result := resolveMonsterCombat(gs)
	if gs.Quests.MonsterHull != MonsterMaxHull {
		t.Error("monster hull should be unchanged when player has no weapons")
	}
	if result.Result == "" {
		t.Error("expected non-empty result message")
	}
}

func TestSpaceMonsterKill(t *testing.T) {
	gs := newTestGameState(t)

	gs.SetQuestState(QuestSpaceMonster, QuestAvailable)
	gs.Quests.MonsterHull = 1

	gs.Player.Skills[formula.SkillFighter] = 10
	gs.Player.Skills[formula.SkillEngineer] = 10

	startCredits := gs.Player.Credits
	startRep := gs.Player.Reputation

	resolveMonsterCombat(gs)

	if gs.QuestState(QuestSpaceMonster) != QuestComplete {
		t.Error("monster with 1 hull should die")
	}
	if gs.Quests.MonsterHull != 0 {
		t.Error("dead monster hull should be 0")
	}
	if gs.Player.Credits != startCredits+10000 {
		t.Errorf("expected +10000 credits bounty, got %d", gs.Player.Credits-startCredits)
	}
	if gs.Player.Reputation != startRep+5 {
		t.Errorf("expected +5 reputation, got %d", gs.Player.Reputation-startRep)
	}
}

func TestScarabHullUpgrade(t *testing.T) {
	gs := newTestGameState(t)

	pulseIdx := findEquipByName(gs, "Pulse Laser")
	if pulseIdx < 0 {
		t.Fatal("Pulse Laser not found in equipment")
	}
	gs.Player.Ship.Weapons = []int{pulseIdx}

	gs.SetQuestState(QuestScarab, QuestAvailable)

	result := resolveQuestChainAction(gs, "Scarab Found!", 0)
	if result.Message == "" && result.Combat == nil {
		t.Fatal("expected non-empty result from scarab attack")
	}

	if gs.QuestState(QuestScarab) != QuestComplete {
		t.Error("expected Scarab quest to be Complete")
	}
	if !gs.Player.Ship.HullUpgraded {
		t.Error("expected HullUpgraded to be true")
	}
	shipDef := gs.Data.Ships[gs.Player.Ship.TypeID]
	expectedHull := shipDef.Hull + ScarabHullBonus
	if gs.Player.Ship.Hull != expectedHull {
		t.Errorf("expected hull %d, got %d", expectedHull, gs.Player.Ship.Hull)
	}
}

func TestScarabNoPulseLaser(t *testing.T) {
	gs := newTestGameState(t)

	gs.Player.Ship.Weapons = []int{}

	gs.SetQuestState(QuestScarab, QuestAvailable)

	result := resolveQuestChainAction(gs, "Scarab Found!", 0)
	if result.Message == "" && result.Combat == nil {
		t.Fatal("expected non-empty result")
	}

	if gs.QuestState(QuestScarab) == QuestComplete {
		t.Error("Scarab should NOT be completed without Pulse Laser")
	}
	if gs.Player.Ship.HullUpgraded {
		t.Error("hull should not be upgraded without Pulse Laser")
	}
}

func TestJarekCompletion(t *testing.T) {
	gs := newTestGameState(t)
	gs.Day = 15

	gs.SetQuestState(QuestJarek, QuestActive)
	gs.SetQuestProgress(QuestJarek, 0)
	gs.Player.Crew = append(gs.Player.Crew, Mercenary{
		Name: "Jarek", Skills: [formula.NumSkills]int{3, 2, 10, 4}, SystemIdx: -1, IsQuest: true,
	})

	devidia := findSystem(gs, "Devidia")
	if devidia < 0 {
		t.Fatal("Devidia not found")
	}
	gs.CurrentSystemID = devidia
	events := CheckQuestsOnArrival(gs)

	if gs.QuestState(QuestJarek) != QuestComplete {
		t.Error("expected Jarek quest Complete")
	}
	if HasQuestCrew(gs, "Jarek") {
		t.Error("expected Jarek removed from crew")
	}

	found := false
	for _, e := range events {
		if e.Title == "Ambassador Delivered!" {
			found = true
		}
	}
	if !found {
		t.Error("expected Ambassador Delivered! event")
	}
}

func TestJarekTimeout(t *testing.T) {
	gs := newTestGameState(t)
	gs.Day = 15

	gs.SetQuestState(QuestJarek, QuestActive)
	gs.SetQuestProgress(QuestJarek, 10)
	gs.Player.Crew = append(gs.Player.Crew, Mercenary{
		Name: "Jarek", Skills: [formula.NumSkills]int{3, 2, 10, 4}, SystemIdx: -1, IsQuest: true,
	})

	otherSys := 0
	devidia := findSystem(gs, "Devidia")
	if otherSys == devidia {
		otherSys = 1
	}
	gs.CurrentSystemID = otherSys
	events := CheckQuestsOnArrival(gs)

	if gs.QuestState(QuestJarek) != QuestUnavailable {
		t.Errorf("expected Jarek quest Unavailable after timeout, got %d", gs.QuestState(QuestJarek))
	}
	if HasQuestCrew(gs, "Jarek") {
		t.Error("expected Jarek removed from crew after timeout")
	}

	found := false
	for _, e := range events {
		if e.Title == "Ambassador Impatient" {
			found = true
		}
	}
	if !found {
		t.Error("expected Ambassador Impatient event")
	}
}

func TestGemulonSuccess(t *testing.T) {
	gs := newTestGameState(t)
	gs.Day = 40

	gs.SetQuestState(QuestGemulon, QuestAvailable)
	gs.SetQuestProgress(QuestGemulon, gs.Day)

	gemulon := findSystem(gs, "Gemulon")
	if gemulon < 0 {
		t.Fatal("Gemulon not found")
	}

	gs.Player.Ship.Gadgets = []int{}

	gs.Day += 3
	gs.CurrentSystemID = gemulon
	events := CheckQuestsOnArrival(gs)

	if gs.QuestState(QuestGemulon) != QuestComplete {
		t.Error("expected Gemulon quest Complete")
	}

	found := false
	for _, e := range events {
		if e.Title == "Gemulon Saved!" {
			found = true
		}
	}
	if !found {
		t.Error("expected Gemulon Saved! event")
	}

	fuelCompactorIdx := findEquipByName(gs, "Fuel Compactor")
	installed := false
	for _, g := range gs.Player.Ship.Gadgets {
		if g == fuelCompactorIdx {
			installed = true
		}
	}
	if !installed && len(gs.Quests.PendingRewards) == 0 {
		t.Error("expected Fuel Compactor installed or in pending rewards")
	}
}

func TestGemulonFailureConsequences(t *testing.T) {
	gs := newTestGameState(t)
	gs.Day = 40

	gs.SetQuestState(QuestGemulon, QuestAvailable)
	gs.SetQuestProgress(QuestGemulon, gs.Day)

	gemulon := findSystem(gs, "Gemulon")
	if gemulon < 0 {
		t.Fatal("Gemulon not found")
	}

	gs.Day += 8

	otherSys := 0
	if otherSys == gemulon {
		otherSys = 1
	}
	gs.CurrentSystemID = otherSys
	events := CheckQuestsOnArrival(gs)

	found := false
	for _, e := range events {
		if e.Title == "Gemulon Invaded" {
			found = true
		}
	}
	if !found {
		t.Error("expected Gemulon Invaded event")
	}

	if gs.Data.Systems[gemulon].TechLevel != gamedata.TechPreAgricultural {
		t.Errorf("expected TechPreAgricultural, got %d", gs.Data.Systems[gemulon].TechLevel)
	}
	if gs.Data.Systems[gemulon].PoliticalSystem != gamedata.PolAnarchy {
		t.Errorf("expected PolAnarchy, got %d", gs.Data.Systems[gemulon].PoliticalSystem)
	}
}

func TestFehlerSingularity(t *testing.T) {
	gs := newTestGameState(t)
	gs.Day = 45

	gs.SetQuestState(QuestFehler, QuestAvailable)
	gs.SetQuestProgress(QuestFehler, gs.Day)

	deneb := findSystem(gs, "Deneb")
	if deneb < 0 {
		t.Fatal("Deneb not found")
	}

	gs.Day += 3
	gs.CurrentSystemID = deneb

	startCredits := gs.Player.Credits
	startRep := gs.Player.Reputation
	events := CheckQuestsOnArrival(gs)

	if gs.QuestState(QuestFehler) != QuestComplete {
		t.Error("expected Fehler quest Complete")
	}
	if !gs.Quests.HasSingularity {
		t.Error("expected HasSingularity to be true")
	}
	if gs.Player.Credits != startCredits+10000 {
		t.Errorf("expected credits %d, got %d", startCredits+10000, gs.Player.Credits)
	}
	if gs.Player.Reputation != startRep+3 {
		t.Errorf("expected rep %d, got %d", startRep+3, gs.Player.Reputation)
	}

	found := false
	for _, e := range events {
		if e.Title == "Experiment Stopped!" {
			found = true
		}
	}
	if !found {
		t.Error("expected Experiment Stopped! event")
	}
}
