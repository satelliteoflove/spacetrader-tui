package game_test

import (
	"testing"

	"github.com/the4ofus/spacetrader-tui/internal/game"
	"github.com/the4ofus/spacetrader-tui/internal/gamedata"
)

func stripShipEquipment(gs *game.GameState) {
	gs.Player.Ship.Weapons = nil
	gs.Player.Ship.Shields = nil
	gs.Player.Ship.Gadgets = nil
}

func findOneOfEachEquipCategory(gs *game.GameState) (weaponID, shieldID, gadgetID int, ok bool) {
	weaponID, shieldID, gadgetID = -1, -1, -1
	for i, eq := range gs.Data.Equipment {
		switch eq.Category {
		case gamedata.EquipWeapon:
			if weaponID == -1 {
				weaponID = i
			}
		case gamedata.EquipShield:
			if shieldID == -1 {
				shieldID = i
			}
		case gamedata.EquipGadget:
			if gadgetID == -1 {
				gadgetID = i
			}
		}
	}
	ok = weaponID >= 0 && shieldID >= 0 && gadgetID >= 0
	return
}

func mostExpensiveShipID(gs *game.GameState) int {
	bestID, bestPrice := 0, 0
	for i, s := range gs.Data.Ships {
		if s.Price > bestPrice {
			bestID, bestPrice = i, s.Price
		}
	}
	return bestID
}

func TestInsurableValueTracksInstalledEquipment(t *testing.T) {
	gs := newTestGame(t)
	stripShipEquipment(gs)
	weaponID, shieldID, gadgetID, ok := findOneOfEachEquipCategory(gs)
	if !ok {
		t.Fatal("test data missing a weapon, shield, or gadget definition")
	}

	baseline := game.InsurableValue(gs)

	gs.Player.Ship.Weapons = []int{weaponID}
	withWeapon := game.InsurableValue(gs)
	if withWeapon <= baseline {
		t.Errorf("weapon install must raise insurable value: baseline=%d, after=%d", baseline, withWeapon)
	}

	gs.Player.Ship.Shields = []int{shieldID}
	withShield := game.InsurableValue(gs)
	if withShield <= withWeapon {
		t.Errorf("shield install must raise insurable value: before=%d, after=%d", withWeapon, withShield)
	}

	gs.Player.Ship.Gadgets = []int{gadgetID}
	withGadget := game.InsurableValue(gs)
	if withGadget <= withShield {
		t.Errorf("gadget install must raise insurable value: before=%d, after=%d", withShield, withGadget)
	}
}

func TestInsurancePremiumScalesWithShipValue(t *testing.T) {
	gs := newTestGame(t)
	stripShipEquipment(gs)
	gs.Player.InsuranceDays = 0

	gs.Player.Ship.TypeID = int(gamedata.ShipFlea)
	cheapPremium := game.InsuranceDailyPremium(gs)

	gs.Player.Ship.TypeID = int(gamedata.ShipWasp)
	expensivePremium := game.InsuranceDailyPremium(gs)

	if expensivePremium <= cheapPremium {
		t.Errorf("expensive ship must cost more to insure than cheap ship: cheap=%d, expensive=%d", cheapPremium, expensivePremium)
	}
}

func TestInsuranceNoClaimDiscountReducesAndCaps(t *testing.T) {
	gs := newTestGame(t)
	stripShipEquipment(gs)
	gs.Player.Ship.TypeID = mostExpensiveShipID(gs)

	gs.Player.InsuranceDays = 0
	day0 := game.InsuranceDailyPremium(gs)
	gs.Player.InsuranceDays = 50
	day50 := game.InsuranceDailyPremium(gs)
	gs.Player.InsuranceDays = 90
	day90 := game.InsuranceDailyPremium(gs)
	gs.Player.InsuranceDays = 500
	day500 := game.InsuranceDailyPremium(gs)

	if !(day0 > day50 && day50 > day90) {
		t.Errorf("premium must decrease with accrued no-claim days: day0=%d, day50=%d, day90=%d", day0, day50, day90)
	}
	if day90 <= 1 {
		t.Fatalf("test setup too cheap: day90 hit the floor (%d), cap check would be masked", day90)
	}
	if day90 != day500 {
		t.Errorf("no-claim discount must cap at day 90: day90=%d, day500=%d", day90, day500)
	}
}
