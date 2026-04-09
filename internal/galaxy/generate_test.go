package galaxy

import (
	"testing"

	"github.com/the4ofus/spacetrader-tui/internal/gamedata"
)

func TestGenerateDeterministic(t *testing.T) {
	systems1 := Generate(12345)
	systems2 := Generate(12345)

	if len(systems1) != len(systems2) {
		t.Fatalf("different lengths: %d vs %d", len(systems1), len(systems2))
	}
	for i := range systems1 {
		if systems1[i].Name != systems2[i].Name {
			t.Errorf("system %d name mismatch: %q vs %q", i, systems1[i].Name, systems2[i].Name)
		}
		if systems1[i].X != systems2[i].X || systems1[i].Y != systems2[i].Y {
			t.Errorf("system %d coords mismatch", i)
		}
		if systems1[i].TechLevel != systems2[i].TechLevel {
			t.Errorf("system %d tech mismatch", i)
		}
		if systems1[i].PoliticalSystem != systems2[i].PoliticalSystem {
			t.Errorf("system %d politics mismatch", i)
		}
	}
}

func TestGenerateCount(t *testing.T) {
	systems := Generate(42)
	if len(systems) != NumSystems {
		t.Errorf("expected %d systems, got %d", NumSystems, len(systems))
	}
}

func TestGenerateQuestNames(t *testing.T) {
	systems := Generate(42)
	names := make(map[string]bool, len(systems))
	for _, s := range systems {
		names[s.Name] = true
	}

	for _, critical := range questCriticalNames {
		if !names[critical] {
			t.Errorf("quest-critical system %q not found", critical)
		}
	}
}

func TestGenerateUniqueNames(t *testing.T) {
	systems := Generate(42)
	seen := make(map[string]bool)
	for _, s := range systems {
		if seen[s.Name] {
			t.Errorf("duplicate system name: %q", s.Name)
		}
		seen[s.Name] = true
	}
}

func TestGenerateMinDistance(t *testing.T) {
	systems := Generate(42)
	for i := range systems {
		for j := i + 1; j < len(systems); j++ {
			d := dist(systems[i].X, systems[i].Y, systems[j].X, systems[j].Y)
			if d < float64(MinDistance)-1 {
				t.Errorf("systems %q and %q too close: %.1f parsecs (min %d)",
					systems[i].Name, systems[j].Name, d, MinDistance)
			}
		}
	}
}

func TestGenerateConnectivity(t *testing.T) {
	for _, seed := range []int64{42, 12345, 99999, 7, 314159} {
		systems := Generate(seed)
		coords := make([][2]int, len(systems))
		for i, s := range systems {
			coords[i] = [2]int{s.X, s.Y}
		}
		components := connectedComponents(coords, MaxShipRange)
		if len(components) != 1 {
			t.Errorf("seed %d: galaxy has %d connected components, want 1", seed, len(components))
		}
	}
}

func TestGenerateValidEnums(t *testing.T) {
	systems := Generate(42)
	for _, sys := range systems {
		if sys.TechLevel < 0 || sys.TechLevel >= gamedata.NumTechLevels {
			t.Errorf("system %q: invalid tech level %d", sys.Name, sys.TechLevel)
		}
		if sys.PoliticalSystem < 0 || sys.PoliticalSystem >= gamedata.NumPoliticalSystems {
			t.Errorf("system %q: invalid political system %d", sys.Name, sys.PoliticalSystem)
		}
		if sys.Resource < 0 || sys.Resource >= gamedata.NumResources {
			t.Errorf("system %q: invalid resource %d", sys.Name, sys.Resource)
		}
		if sys.Size < 0 || sys.Size >= gamedata.NumSystemSizes {
			t.Errorf("system %q: invalid size %d", sys.Name, sys.Size)
		}
	}
}

func TestGenerateDifferentSeeds(t *testing.T) {
	systems1 := Generate(1)
	systems2 := Generate(2)

	differences := 0
	for i := range systems1 {
		if systems1[i].Name != systems2[i].Name {
			differences++
		}
	}
	if differences == 0 {
		t.Error("different seeds produced identical name assignments")
	}
}

func TestFindSystemByName(t *testing.T) {
	systems := Generate(42)
	idx := FindSystemByName(systems, "Acamar")
	if idx < 0 {
		t.Fatal("Acamar not found")
	}
	if systems[idx].Name != "Acamar" {
		t.Errorf("got %q, want Acamar", systems[idx].Name)
	}

	idx = FindSystemByName(systems, "NonExistent")
	if idx != -1 {
		t.Errorf("expected -1 for non-existent, got %d", idx)
	}
}
