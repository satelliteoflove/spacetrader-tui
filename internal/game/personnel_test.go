package game

import (
	"testing"

	"github.com/the4ofus/spacetrader-tui/internal/formula"
)

func TestMercenaryWageFormula(t *testing.T) {
	m := Mercenary{Name: "Test", Skills: [formula.NumSkills]int{5, 5, 5, 5}}
	if m.Wage() != 60 {
		t.Errorf("expected wage 60, got %d", m.Wage())
	}

	m2 := Mercenary{Name: "Low", Skills: [formula.NumSkills]int{2, 2, 2, 2}}
	if m2.Wage() != 24 {
		t.Errorf("expected wage 24, got %d", m2.Wage())
	}

	quest := Mercenary{Name: "Jarek", Skills: [formula.NumSkills]int{3, 2, 10, 4}, IsQuest: true}
	if quest.Wage() != 0 {
		t.Errorf("expected quest crew wage 0, got %d", quest.Wage())
	}
}

func TestPersistentMercenaryPool(t *testing.T) {
	gs := newTestGameState(t)

	if len(gs.Mercenaries) != 29 {
		t.Errorf("expected 29 mercenaries, got %d", len(gs.Mercenaries))
	}

	names := map[string]bool{}
	for _, m := range gs.Mercenaries {
		if names[m.Name] {
			t.Errorf("duplicate mercenary name: %s", m.Name)
		}
		names[m.Name] = true
		if m.SystemIdx < 0 || m.SystemIdx >= len(gs.Data.Systems) {
			t.Errorf("merc %s has invalid system %d", m.Name, m.SystemIdx)
		}
		for _, s := range m.Skills {
			if s < 1 || s > 10 {
				t.Errorf("merc %s has out-of-range skill %d (expected 1-10)", m.Name, s)
			}
		}
	}
}

func TestMercenarySystemLimit(t *testing.T) {
	gs := newTestGameState(t)
	counts := map[int]int{}
	for _, m := range gs.Mercenaries {
		counts[m.SystemIdx]++
	}
	for sys, count := range counts {
		if count > 3 {
			t.Errorf("system %d has %d mercenaries (max 3)", sys, count)
		}
	}
}

func TestAvailableMercenariesAtSystem(t *testing.T) {
	gs := newTestGameState(t)

	gs.Mercenaries[0].SystemIdx = gs.CurrentSystemID
	gs.Mercenaries[1].SystemIdx = gs.CurrentSystemID

	available := AvailableMercenaries(gs)
	if len(available) < 2 {
		t.Errorf("expected at least 2 available, got %d", len(available))
	}

	for _, idx := range available {
		m := gs.Mercenaries[idx]
		if m.SystemIdx != gs.CurrentSystemID {
			t.Errorf("merc %s not at current system", m.Name)
		}
	}
}

func TestHireMercenary(t *testing.T) {
	gs := newTestGameState(t)

	shipDef := gs.Data.Ships[gs.Player.Ship.TypeID]
	if shipDef.CrewQuarters <= 1 {
		gs.Player.Ship.TypeID = 4
	}

	gs.Mercenaries[0].SystemIdx = gs.CurrentSystemID
	startCredits := gs.Player.Credits

	ok, _ := HireMercenary(gs, 0)
	if !ok {
		t.Fatal("expected successful hire")
	}
	if len(gs.Player.Crew) != 1 {
		t.Errorf("expected 1 crew member, got %d", len(gs.Player.Crew))
	}
	if gs.Mercenaries[0].SystemIdx != -1 {
		t.Error("expected merc system set to -1 after hire")
	}
	if gs.Player.Credits != startCredits {
		t.Error("hiring should not cost credits (no signing bonus in original)")
	}
}

func TestHireNoQuarters(t *testing.T) {
	gs := newTestGameState(t)

	gs.Mercenaries[0].SystemIdx = gs.CurrentSystemID

	ok, msg := HireMercenary(gs, 0)
	if ok {
		t.Error("should not hire when no quarters available")
	}
	if msg != "No crew quarters available." {
		t.Errorf("unexpected message: %s", msg)
	}
}

func TestFireMercenary(t *testing.T) {
	gs := newTestGameState(t)
	gs.Player.Ship.TypeID = 4

	gs.Mercenaries[0].SystemIdx = gs.CurrentSystemID
	HireMercenary(gs, 0)

	ok, _ := FireMercenary(gs, 0)
	if !ok {
		t.Fatal("expected successful fire")
	}
	if len(gs.Player.Crew) != 0 {
		t.Errorf("expected 0 crew after fire, got %d", len(gs.Player.Crew))
	}
	if gs.Mercenaries[0].SystemIdx == -1 {
		t.Error("expected merc reassigned to a system after fire")
	}
}

func TestFireQuestCrewBlocked(t *testing.T) {
	gs := newTestGameState(t)
	gs.Player.Crew = append(gs.Player.Crew, Mercenary{
		Name: "Jarek", Skills: [formula.NumSkills]int{3, 2, 10, 4}, SystemIdx: -1, IsQuest: true,
	})

	ok, _ := FireMercenary(gs, 0)
	if ok {
		t.Error("should not be able to fire quest crew")
	}
	if len(gs.Player.Crew) != 1 {
		t.Error("quest crew should still be on board")
	}
}

func TestQuestCrewContributesSkills(t *testing.T) {
	gs := newTestGameState(t)
	gs.Player.Skills = [formula.NumSkills]int{3, 3, 3, 3}

	baseFighter := EffectivePlayerSkill(gs, formula.SkillFighter)

	gs.Player.Crew = append(gs.Player.Crew, Mercenary{
		Name: "Wild", Skills: WildSkills, SystemIdx: -1, IsQuest: true,
	})

	withWild := EffectivePlayerSkill(gs, formula.SkillFighter)
	if withWild <= baseFighter {
		t.Errorf("Wild's fighter skill should boost effective skill: base=%d, with=%d", baseFighter, withWild)
	}
}

func TestHagglingComputerBonus(t *testing.T) {
	gs := newTestGameState(t)
	gs.Player.Skills = [formula.NumSkills]int{5, 5, 5, 5}

	baseTrade := EffectivePlayerSkill(gs, formula.SkillTrader)

	gs.SetQuestState(QuestJarek, QuestComplete)

	withHaggle := EffectivePlayerSkill(gs, formula.SkillTrader)
	if withHaggle != baseTrade+1 {
		t.Errorf("haggling computer should add +1 trader: base=%d, with=%d", baseTrade, withHaggle)
	}
}

func TestNthLowestSkill(t *testing.T) {
	skills := [formula.NumSkills]int{5, 3, 8, 1}
	lowest := NthLowestSkill(skills, 1)
	if lowest != 3 {
		t.Errorf("expected skill index 3 (engineer=1) as lowest, got %d", lowest)
	}
	secondLowest := NthLowestSkill(skills, 2)
	if secondLowest != 1 {
		t.Errorf("expected skill index 1 (fighter=3) as second lowest, got %d", secondLowest)
	}
}

func TestZeethibalCreation(t *testing.T) {
	gs := newTestGameState(t)
	gs.Player.Skills = [formula.NumSkills]int{5, 3, 8, 1}

	CreateZeethibal(gs)

	var zeet *Mercenary
	for i, m := range gs.Mercenaries {
		if m.Name == "Zeethibal" {
			zeet = &gs.Mercenaries[i]
			break
		}
	}
	if zeet == nil {
		t.Fatal("Zeethibal not found in mercenary pool")
	}

	kravat := findSystem(gs, "Kravat")
	if zeet.SystemIdx != kravat {
		t.Errorf("expected Zeethibal at Kravat (%d), got %d", kravat, zeet.SystemIdx)
	}
	if !zeet.IsQuest {
		t.Error("expected Zeethibal to be quest crew (free)")
	}
	if zeet.Skills[3] != 10 {
		t.Errorf("expected 10 in engineer (worst skill), got %d", zeet.Skills[3])
	}
	if zeet.Skills[1] != 8 {
		t.Errorf("expected 8 in fighter (second worst), got %d", zeet.Skills[1])
	}
	if zeet.Skills[0] != 5 || zeet.Skills[2] != 5 {
		t.Errorf("expected 5 in other skills, got pilot=%d trader=%d", zeet.Skills[0], zeet.Skills[2])
	}
}

func TestClearCrewResetsQuests(t *testing.T) {
	gs := newTestGameState(t)

	gs.SetQuestState(QuestJarek, QuestActive)
	gs.Player.Crew = append(gs.Player.Crew, Mercenary{
		Name: "Jarek", Skills: [formula.NumSkills]int{3, 2, 10, 4}, SystemIdx: -1, IsQuest: true,
	})

	ClearCrewAndResetQuests(gs)

	if gs.QuestState(QuestJarek) != QuestUnavailable {
		t.Errorf("expected Jarek quest reset to Unavailable, got %d", gs.QuestState(QuestJarek))
	}
	if len(gs.Player.Crew) != 0 {
		t.Error("expected crew cleared")
	}
}
