package shipyard_test

import (
	"os"
	"testing"

	"github.com/the4ofus/spacetrader-tui/internal/data"
	"github.com/the4ofus/spacetrader-tui/internal/formula"
	"github.com/the4ofus/spacetrader-tui/internal/game"
	"github.com/the4ofus/spacetrader-tui/internal/gamedata"
	"github.com/the4ofus/spacetrader-tui/internal/shipyard"
)

func newTestGame(t *testing.T) *game.GameState {
	t.Helper()
	gd, err := data.LoadAll(os.DirFS("../../data"))
	if err != nil {
		t.Fatalf("LoadAll: %v", err)
	}
	skills := [formula.NumSkills]int{4, 4, 4, 4}
	gs := game.NewGameWithSeed(gd, "Test", skills, gamedata.DiffNormal, 42)
	for i, sys := range gd.Systems {
		if sys.TechLevel >= gamedata.TechIndustrial {
			gs.CurrentSystemID = i
			break
		}
	}
	return gs
}

func TestAvailableShips(t *testing.T) {
	gs := newTestGame(t)
	ships := shipyard.AvailableShips(gs)
	if len(ships) == 0 {
		t.Error("no ships available")
	}

	sys := gs.Data.Systems[gs.CurrentSystemID]
	for _, s := range ships {
		if s.MinTech > sys.TechLevel {
			t.Errorf("ship %q requires tech %v but system is %v", s.Name, s.MinTech, sys.TechLevel)
		}
	}
}

func TestAvailableEquipment(t *testing.T) {
	gs := newTestGame(t)
	equip := shipyard.AvailableEquipment(gs)
	if len(equip) == 0 {
		t.Error("no equipment available")
	}
}

func TestShipHullTradeIn(t *testing.T) {
	gs := newTestGame(t)
	value := shipyard.ShipHullTradeIn(gs)

	gnatPrice := gs.Data.Ships[int(gamedata.ShipGnat)].Price
	expected := gnatPrice * 3 / 4
	if value != expected {
		t.Errorf("hull trade-in: got %d, want %d", value, expected)
	}
}

func TestBuyShip(t *testing.T) {
	gs := newTestGame(t)
	gs.Player.Credits = 100000

	available := shipyard.AvailableShips(gs)
	if len(available) < 2 {
		t.Skip("not enough ships available at this tech level")
	}

	var target gamedata.ShipDef
	for _, s := range available {
		if s.ID != gs.Player.Ship.TypeID {
			target = s
			break
		}
	}

	result := shipyard.BuyShip(gs, target.ID, nil)
	if !result.Success {
		t.Fatalf("BuyShip failed: %s", result.Message)
	}
	if gs.Player.Ship.TypeID != target.ID {
		t.Errorf("ship type: got %d, want %d", gs.Player.Ship.TypeID, target.ID)
	}
}

func TestBuyShipInsufficientCredits(t *testing.T) {
	gs := newTestGame(t)
	gs.Player.Credits = 0

	result := shipyard.BuyShip(gs, int(gamedata.ShipWasp), nil)
	if result.Success {
		t.Error("should fail with 0 credits for Wasp")
	}
}

func TestBuyShipTransfersEquipment(t *testing.T) {
	gs := newTestGame(t)
	gs.Player.Credits = 200000

	gs.Player.Ship.Weapons = []int{0}

	fireflyDef := gs.Data.Ships[int(gamedata.ShipFirefly)]
	if fireflyDef.WeaponSlots < 1 {
		t.Skip("Firefly has no weapon slots")
	}

	result := shipyard.BuyShip(gs, int(gamedata.ShipFirefly), nil)
	if !result.Success {
		t.Fatalf("BuyShip failed: %s", result.Message)
	}
	if len(gs.Player.Ship.Weapons) != 1 {
		t.Errorf("weapons: got %d, want 1 (pulse laser should transfer)", len(gs.Player.Ship.Weapons))
	}
	if gs.Player.Ship.Weapons[0] != 0 {
		t.Errorf("weapon ID: got %d, want 0 (pulse laser)", gs.Player.Ship.Weapons[0])
	}
}

func TestBuyShipSellsExcessEquipment(t *testing.T) {
	gs := newTestGame(t)
	gs.Player.Credits = 200000

	gs.Player.Ship.TypeID = int(gamedata.ShipHornet)
	gs.Player.Ship.Weapons = []int{0, 1, 2}

	creditsBefore := gs.Player.Credits

	fireflyDef := gs.Data.Ships[int(gamedata.ShipFirefly)]
	if fireflyDef.WeaponSlots >= 3 {
		t.Skip("Firefly has enough weapon slots for all 3 weapons")
	}

	preview := shipyard.PreviewShipPurchase(gs, int(gamedata.ShipFirefly))
	if preview.Error != "" {
		t.Fatalf("preview error: %s", preview.Error)
	}

	if len(preview.Weapons.Kept) != fireflyDef.WeaponSlots {
		t.Errorf("kept weapons: got %d, want %d", len(preview.Weapons.Kept), fireflyDef.WeaponSlots)
	}
	expectedSold := 3 - fireflyDef.WeaponSlots
	if len(preview.Weapons.Sold) != expectedSold {
		t.Errorf("sold weapons: got %d, want %d", len(preview.Weapons.Sold), expectedSold)
	}

	result := shipyard.BuyShip(gs, int(gamedata.ShipFirefly), nil)
	if !result.Success {
		t.Fatalf("BuyShip failed: %s", result.Message)
	}
	if len(gs.Player.Ship.Weapons) != fireflyDef.WeaponSlots {
		t.Errorf("weapons after: got %d, want %d", len(gs.Player.Ship.Weapons), fireflyDef.WeaponSlots)
	}

	expectedCredits := creditsBefore - preview.NetCost
	if gs.Player.Credits != expectedCredits {
		t.Errorf("credits: got %d, want %d", gs.Player.Credits, expectedCredits)
	}
}

func TestBuyShipKeepsMostValuable(t *testing.T) {
	gs := newTestGame(t)
	gs.Player.Credits = 200000

	gs.Player.Ship.TypeID = int(gamedata.ShipHornet)
	gs.Player.Ship.Weapons = []int{0, 1, 2}

	fireflyDef := gs.Data.Ships[int(gamedata.ShipFirefly)]
	if fireflyDef.WeaponSlots != 1 {
		t.Skipf("Firefly has %d weapon slots, test expects 1", fireflyDef.WeaponSlots)
	}

	preview := shipyard.PreviewShipPurchase(gs, int(gamedata.ShipFirefly))
	if preview.Error != "" {
		t.Fatalf("preview error: %s", preview.Error)
	}

	militaryLaserID := 2
	if len(preview.Weapons.Kept) != 1 || preview.Weapons.Kept[0] != militaryLaserID {
		t.Errorf("should keep most expensive weapon (Military Laser, ID 2): got kept=%v", preview.Weapons.Kept)
	}
}

func TestBuyEquipmentSlotLimit(t *testing.T) {
	gs := newTestGame(t)
	gs.Player.Credits = 100000

	gs.Player.Ship.Weapons = []int{0}

	result := shipyard.BuyEquipment(gs, 1)
	if result.Success {
		t.Error("Gnat has 1 weapon slot, should fail with slot full")
	}
}

func TestSellEquipment(t *testing.T) {
	gs := newTestGame(t)
	startCredits := gs.Player.Credits

	result := shipyard.SellEquipment(gs, gamedata.EquipWeapon, 0)
	if !result.Success {
		t.Fatalf("SellEquipment failed: %s", result.Message)
	}
	if gs.Player.Credits <= startCredits {
		t.Error("should gain credits from selling")
	}
	if len(gs.Player.Ship.Weapons) != 0 {
		t.Error("weapon slot should be empty")
	}
}

func TestRepair(t *testing.T) {
	gs := newTestGame(t)
	gs.Player.Ship.Hull = 50
	gs.Player.Credits = 10000

	result := shipyard.Repair(gs)
	if !result.Success {
		t.Fatalf("Repair failed: %s", result.Message)
	}
	if gs.Player.Ship.Hull != gs.Data.Ships[gs.Player.Ship.TypeID].Hull {
		t.Error("hull should be fully restored")
	}
}

func TestRepairAlreadyFull(t *testing.T) {
	gs := newTestGame(t)
	result := shipyard.Repair(gs)
	if result.Success {
		t.Error("should report already fully repaired")
	}
}

func TestRefuel(t *testing.T) {
	gs := newTestGame(t)
	gs.Player.Ship.Fuel = 5
	gs.Player.Credits = 10000

	result := shipyard.Refuel(gs)
	if !result.Success {
		t.Fatalf("Refuel failed: %s", result.Message)
	}
	if gs.Player.Ship.Fuel != gs.Data.Ships[gs.Player.Ship.TypeID].Range {
		t.Error("fuel should be full")
	}
}

func TestRefuelAlreadyFull(t *testing.T) {
	gs := newTestGame(t)
	result := shipyard.Refuel(gs)
	if result.Success {
		t.Error("should report tank already full")
	}
}
