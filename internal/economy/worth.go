package economy

import "github.com/the4ofus/spacetrader-tui/internal/game"

func PlayerWorth(gs *game.GameState) int {
	worth := gs.Player.Credits

	shipDef := gs.Data.Ships[gs.Player.Ship.TypeID]
	worth += shipDef.Price * 3 / 4

	for _, wID := range gs.Player.Ship.Weapons {
		worth += gs.Data.Equipment[wID].Price * 2 / 3
	}
	for _, sID := range gs.Player.Ship.Shields {
		worth += gs.Data.Equipment[sID].Price * 2 / 3
	}
	for _, gID := range gs.Player.Ship.Gadgets {
		worth += gs.Data.Equipment[gID].Price * 2 / 3
	}

	worth -= gs.Player.LoanBalance

	return worth
}
