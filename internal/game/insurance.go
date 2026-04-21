package game

const MaxInsuranceNoClaimDiscount = 90

func InsurableValue(gs *GameState) int {
	total := gs.ShipDef(gs.Player.Ship.TypeID).Price
	for _, eid := range gs.Player.Ship.Weapons {
		total += gs.EquipDef(eid).Price
	}
	for _, eid := range gs.Player.Ship.Shields {
		total += gs.EquipDef(eid).Price
	}
	for _, eid := range gs.Player.Ship.Gadgets {
		total += gs.EquipDef(eid).Price
	}
	return total
}

func InsuranceBasePremium(gs *GameState) int {
	base := InsurableValue(gs) / 1000
	if base < 1 {
		base = 1
	}
	return base
}

func InsuranceNoClaimDiscount(gs *GameState) int {
	d := gs.Player.InsuranceDays
	if d > MaxInsuranceNoClaimDiscount {
		d = MaxInsuranceNoClaimDiscount
	}
	return d
}

func InsuranceDailyPremium(gs *GameState) int {
	base := InsuranceBasePremium(gs)
	discount := InsuranceNoClaimDiscount(gs)
	premium := base * (100 - discount) / 100
	if premium < 1 {
		premium = 1
	}
	return premium
}
