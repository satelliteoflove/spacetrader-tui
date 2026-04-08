package game

import "fmt"

type CargoResult struct {
	Success bool
	Message string
}

const DumpCostBase = 5

func validateCargoOp(gs *GameState, goodIdx, qty int) *CargoResult {
	if goodIdx < 0 || goodIdx >= NumGoods || qty <= 0 {
		return &CargoResult{Message: "Invalid request."}
	}
	if gs.Player.Cargo[goodIdx] < qty {
		return &CargoResult{Message: "Not enough cargo."}
	}
	return nil
}

func DumpCargo(gs *GameState, goodIdx int, qty int) CargoResult {
	if err := validateCargoOp(gs, goodIdx, qty); err != nil {
		return *err
	}

	costPerUnit := DumpCostBase * (int(gs.Difficulty) + 1)
	totalCost := costPerUnit * qty
	if gs.Player.Credits < totalCost {
		return CargoResult{Message: fmt.Sprintf("Dumping costs %d cr (%d cr/unit). Not enough credits.", totalCost, costPerUnit)}
	}

	gs.Player.Credits -= totalCost
	gs.Player.Cargo[goodIdx] -= qty
	if gs.Player.Cargo[goodIdx] == 0 {
		gs.Player.CargoCost[goodIdx] = 0
	}

	goodName := gs.Data.Goods[goodIdx].Name
	return CargoResult{
		Success: true,
		Message: fmt.Sprintf("Dumped %d %s for %d cr disposal fee.", qty, goodName, totalCost),
	}
}

func JettisonCargo(gs *GameState, goodIdx int, qty int) CargoResult {
	if err := validateCargoOp(gs, goodIdx, qty); err != nil {
		return *err
	}

	gs.Player.Cargo[goodIdx] -= qty
	if gs.Player.Cargo[goodIdx] == 0 {
		gs.Player.CargoCost[goodIdx] = 0
	}
	gs.Player.PoliceRecord--

	goodName := gs.Data.Goods[goodIdx].Name
	return CargoResult{
		Success: true,
		Message: fmt.Sprintf("Jettisoned %d %s. Littering is illegal -- police record worsened.", qty, goodName),
	}
}
