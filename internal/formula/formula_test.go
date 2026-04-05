package formula

import (
	"math"
	"testing"

	"github.com/the4ofus/spacetrader-tui/internal/gamedata"
)

func TestDistance(t *testing.T) {
	tests := []struct {
		x1, y1, x2, y2 int
		want            float64
	}{
		{0, 0, 3, 4, 5.0},
		{0, 0, 0, 0, 0.0},
		{10, 10, 10, 10, 0.0},
		{0, 0, 1, 0, 1.0},
		{20, 34, 12, 23, math.Sqrt(64 + 121)},
	}

	for _, tt := range tests {
		got := Distance(tt.x1, tt.y1, tt.x2, tt.y2)
		if math.Abs(got-tt.want) > 0.001 {
			t.Errorf("Distance(%d,%d,%d,%d) = %f, want %f",
				tt.x1, tt.y1, tt.x2, tt.y2, got, tt.want)
		}
	}
}

func TestSkillPointsForDifficulty(t *testing.T) {
	tests := []struct {
		diff gamedata.Difficulty
		want int
	}{
		{gamedata.DiffBeginner, 20},
		{gamedata.DiffEasy, 18},
		{gamedata.DiffNormal, 16},
		{gamedata.DiffHard, 14},
		{gamedata.DiffImpossible, 12},
	}

	for _, tt := range tests {
		got := SkillPointsForDifficulty(tt.diff)
		if got != tt.want {
			t.Errorf("SkillPointsForDifficulty(%v) = %d, want %d", tt.diff, got, tt.want)
		}
	}
}

type testMerc struct {
	skills [NumSkills]int
}

func (m testMerc) GetSkills() [NumSkills]int { return m.skills }

func TestEffectiveSkill(t *testing.T) {
	crew := []Mercenary{
		testMerc{skills: [4]int{3, 8, 2, 1}},
		testMerc{skills: [4]int{5, 2, 7, 4}},
	}

	if got := EffectiveSkill(4, crew, SkillPilot, 0); got != 5 {
		t.Errorf("pilot skill: got %d, want 5 (best merc)", got)
	}
	if got := EffectiveSkill(4, crew, SkillFighter, 0); got != 8 {
		t.Errorf("fighter skill: got %d, want 8 (best merc)", got)
	}
	if got := EffectiveSkill(4, crew, SkillTrader, 1); got != 8 {
		t.Errorf("trader skill: got %d, want 8 (best merc 7 + gadget 1)", got)
	}
	if got := EffectiveSkill(9, nil, SkillPilot, 0); got != 9 {
		t.Errorf("no crew: got %d, want 9 (player only)", got)
	}
}

func TestPoliceRecordTiers(t *testing.T) {
	tests := []struct {
		record int
		want   gamedata.PoliceRecordTier
	}{
		{-150, gamedata.RecordPsychopath},
		{-80, gamedata.RecordVillain},
		{-50, gamedata.RecordCriminal},
		{-15, gamedata.RecordCrook},
		{-5, gamedata.RecordDubious},
		{0, gamedata.RecordClean},
		{15, gamedata.RecordLawful},
		{50, gamedata.RecordTrusted},
		{80, gamedata.RecordLiked},
		{150, gamedata.RecordHero},
	}

	for _, tt := range tests {
		got := gamedata.PoliceRecordToTier(tt.record)
		if got != tt.want {
			t.Errorf("PoliceRecordToTier(%d) = %v, want %v", tt.record, got, tt.want)
		}
	}
}

func TestReputationTiers(t *testing.T) {
	tests := []struct {
		rep  int
		want gamedata.ReputationTier
	}{
		{0, gamedata.RepHarmless},
		{1, gamedata.RepMostlyHarmless},
		{5, gamedata.RepPoor},
		{10, gamedata.RepAverage},
		{20, gamedata.RepAboveAverage},
		{30, gamedata.RepCompetent},
		{60, gamedata.RepDangerous},
		{150, gamedata.RepDeadly},
		{200, gamedata.RepElite},
	}

	for _, tt := range tests {
		got := gamedata.ReputationToTier(tt.rep)
		if got != tt.want {
			t.Errorf("ReputationToTier(%d) = %v, want %v", tt.rep, got, tt.want)
		}
	}
}
