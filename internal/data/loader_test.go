package data_test

import (
	"os"
	"testing"

	"github.com/the4ofus/spacetrader-tui/internal/data"
	"github.com/the4ofus/spacetrader-tui/internal/gamedata"
)

func loadTestData(t *testing.T) *gamedata.GameData {
	t.Helper()
	dataFS := os.DirFS("../../data")
	gd, err := data.LoadAll(dataFS)
	if err != nil {
		t.Fatalf("LoadAll failed: %v", err)
	}
	return gd
}

func TestLoadAllCounts(t *testing.T) {
	gd := loadTestData(t)

	if got := len(gd.Goods); got != 10 {
		t.Errorf("expected 10 goods, got %d", got)
	}
	if got := len(gd.Ships); got != 10 {
		t.Errorf("expected 10 ships, got %d", got)
	}
	if got := len(gd.Equipment); got != 14 {
		t.Errorf("expected 14 equipment (11 base + 3 quest), got %d", got)
	}
}

func TestGoodEnumConversions(t *testing.T) {
	gd := loadTestData(t)

	for _, g := range gd.Goods {
		if g.MinTech < 0 || g.MinTech >= gamedata.NumTechLevels {
			t.Errorf("good %q: invalid min tech %d", g.Name, g.MinTech)
		}
		if g.MaxTech < 0 || g.MaxTech >= gamedata.NumTechLevels {
			t.Errorf("good %q: invalid max tech %d", g.Name, g.MaxTech)
		}
		if g.MinTech > g.MaxTech {
			t.Errorf("good %q: min tech %d > max tech %d", g.Name, g.MinTech, g.MaxTech)
		}
		if g.BasePrice <= 0 {
			t.Errorf("good %q: invalid base price %d", g.Name, g.BasePrice)
		}
	}
}

func TestShipEnumConversions(t *testing.T) {
	gd := loadTestData(t)

	for _, s := range gd.Ships {
		if s.MinTech < 0 || s.MinTech >= gamedata.NumTechLevels {
			t.Errorf("ship %q: invalid min tech %d", s.Name, s.MinTech)
		}
		if s.Hull <= 0 {
			t.Errorf("ship %q: invalid hull %d", s.Name, s.Hull)
		}
		if s.Price <= 0 {
			t.Errorf("ship %q: invalid price %d", s.Name, s.Price)
		}
	}
}

func TestEquipmentCategories(t *testing.T) {
	gd := loadTestData(t)

	weapons, shields, gadgets := 0, 0, 0
	for _, e := range gd.Equipment {
		switch e.Category {
		case gamedata.EquipWeapon:
			weapons++
			if e.Power <= 0 {
				t.Errorf("weapon %q: invalid power %d", e.Name, e.Power)
			}
		case gamedata.EquipShield:
			shields++
			if e.Protection <= 0 {
				t.Errorf("shield %q: invalid protection %d", e.Name, e.Protection)
			}
		case gamedata.EquipGadget:
			gadgets++
		}
	}

	if weapons != 4 {
		t.Errorf("expected 4 weapons (3 base + Morgan's Laser), got %d", weapons)
	}
	if shields != 3 {
		t.Errorf("expected 3 shields (2 base + Lightning Shield), got %d", shields)
	}
	if gadgets != 7 {
		t.Errorf("expected 7 gadgets (5 base + Trade Analyzer + Fuel Compactor), got %d", gadgets)
	}
}

func TestGoodIDs(t *testing.T) {
	gd := loadTestData(t)

	expected := map[string]gamedata.GoodID{
		"Water":    gamedata.GoodWater,
		"Firearms": gamedata.GoodFirearms,
		"Narcotics": gamedata.GoodNarcotics,
		"Robots":   gamedata.GoodRobots,
	}

	for _, g := range gd.Goods {
		if want, ok := expected[g.Name]; ok {
			if g.ID != want {
				t.Errorf("good %q: got ID %d, want %d", g.Name, g.ID, want)
			}
		}
	}
}

func TestIllegalGoods(t *testing.T) {
	gd := loadTestData(t)

	illegal := 0
	for _, g := range gd.Goods {
		if !g.Legal {
			illegal++
			if g.Name != "Firearms" && g.Name != "Narcotics" {
				t.Errorf("unexpected illegal good: %q", g.Name)
			}
		}
	}
	if illegal != 2 {
		t.Errorf("expected 2 illegal goods, got %d", illegal)
	}
}
