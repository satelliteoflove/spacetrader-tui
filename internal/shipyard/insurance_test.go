package shipyard_test

import (
	"testing"

	"github.com/the4ofus/spacetrader-tui/internal/gamedata"
	"github.com/the4ofus/spacetrader-tui/internal/shipyard"
)

func TestBuyInsuranceCostReflectsShipValue(t *testing.T) {
	setup := func(shipType gamedata.ShipType) int {
		gs := newTestGame(t)
		gs.Player.HasEscapePod = true
		gs.Player.Credits = 100000
		gs.Player.Ship.TypeID = int(shipType)
		gs.Player.Ship.Weapons = nil
		gs.Player.Ship.Shields = nil
		gs.Player.Ship.Gadgets = nil
		before := gs.Player.Credits
		r := shipyard.BuyInsurance(gs)
		if !r.Success {
			t.Fatalf("BuyInsurance failed for %v: %s", shipType, r.Message)
		}
		return before - gs.Player.Credits
	}

	cheapCost := setup(gamedata.ShipFlea)
	expensiveCost := setup(gamedata.ShipWasp)

	if expensiveCost <= cheapCost {
		t.Errorf("Wasp insurance must cost more than Flea: flea=%d, wasp=%d", cheapCost, expensiveCost)
	}
}
