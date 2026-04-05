package market_test

import (
	"math/rand"
	"os"
	"testing"

	"github.com/the4ofus/spacetrader-tui/internal/data"
	"github.com/the4ofus/spacetrader-tui/internal/formula"
	"github.com/the4ofus/spacetrader-tui/internal/game"
	"github.com/the4ofus/spacetrader-tui/internal/gamedata"
	"github.com/the4ofus/spacetrader-tui/internal/market"
)

func loadData(t *testing.T) *gamedata.GameData {
	t.Helper()
	gd, err := data.LoadAll(os.DirFS("../../data"))
	if err != nil {
		t.Fatalf("LoadAll: %v", err)
	}
	return gd
}

func newTestGame(t *testing.T) *game.GameState {
	t.Helper()
	gd := loadData(t)
	skills := [formula.NumSkills]int{4, 4, 4, 4}
	return game.NewGameWithSeed(gd, "Test", skills, gamedata.DiffNormal, 42)
}

func TestCalculatePriceBasic(t *testing.T) {
	gd := loadData(t)
	rng := rand.New(rand.NewSource(99))

	water := gd.Goods[int(gamedata.GoodWater)]

	for _, sys := range gd.Systems {
		price := market.CalculatePrice(water, sys, "", 0, rng)
		if price < 0 {
			continue
		}
		if price < 1 {
			t.Errorf("system %q: water price %d < 1", sys.Name, price)
		}
		if price > 200 {
			t.Errorf("system %q: water price %d seems too high (base 30)", sys.Name, price)
		}
	}
}

func TestCalculatePriceUnavailable(t *testing.T) {
	gd := loadData(t)

	robots := gd.Goods[int(gamedata.GoodRobots)]

	for _, sys := range gd.Systems {
		if sys.TechLevel < gamedata.TechPostIndustrial {
			price := market.CalculatePrice(robots, sys, "", 0, nil)
			if price != -1 {
				t.Errorf("system %q (tech %v): robots should be unavailable, got price %d",
					sys.Name, sys.TechLevel, price)
			}
		}
	}
}

func TestCalculatePriceTraderDiscount(t *testing.T) {
	gd := loadData(t)

	water := gd.Goods[int(gamedata.GoodWater)]
	sys := gd.Systems[0]

	priceNoSkill := market.CalculatePrice(water, sys, "", 0, nil)
	priceSkill10 := market.CalculatePrice(water, sys, "", 10, nil)

	if priceSkill10 >= priceNoSkill {
		t.Errorf("trader skill 10 should reduce price: no skill=%d, skill 10=%d",
			priceNoSkill, priceSkill10)
	}
}

func TestCalculatePriceEvent(t *testing.T) {
	gd := loadData(t)

	water := gd.Goods[int(gamedata.GoodWater)]
	sys := gd.Systems[0]

	priceNormal := market.CalculatePrice(water, sys, "", 0, nil)
	priceDrought := market.CalculatePrice(water, sys, "Drought", 0, nil)

	if priceDrought <= priceNormal {
		t.Errorf("drought should increase water price: normal=%d, drought=%d",
			priceNormal, priceDrought)
	}
}

func TestBuySuccess(t *testing.T) {
	gs := newTestGame(t)

	var waterIdx int
	for i := range gs.Systems[gs.CurrentSystemID].Prices {
		if gs.Systems[gs.CurrentSystemID].Prices[i] > 0 {
			waterIdx = i
			break
		}
	}

	startCredits := gs.Player.Credits
	result := market.Buy(gs, waterIdx, 1)
	if !result.Success {
		t.Fatalf("buy failed: %s", result.Message)
	}
	if gs.Player.Credits >= startCredits {
		t.Error("credits should have decreased")
	}
	if gs.Player.Cargo[waterIdx] != 1 {
		t.Errorf("cargo: got %d, want 1", gs.Player.Cargo[waterIdx])
	}
}

func TestBuyInsufficientCredits(t *testing.T) {
	gs := newTestGame(t)
	gs.Player.Credits = 0

	result := market.Buy(gs, 0, 1)
	if result.Success {
		t.Error("should fail with 0 credits")
	}
}

func TestBuyInsufficientCargo(t *testing.T) {
	gs := newTestGame(t)
	gs.Player.Credits = 999999

	gs.Player.Cargo[0] = 15

	result := market.Buy(gs, 1, 1)
	if result.Success {
		t.Error("should fail with full cargo")
	}
}

func TestSellSuccess(t *testing.T) {
	gs := newTestGame(t)

	var goodIdx int
	for i := range gs.Systems[gs.CurrentSystemID].Prices {
		if gs.Systems[gs.CurrentSystemID].Prices[i] > 0 {
			goodIdx = i
			break
		}
	}

	gs.Player.Cargo[goodIdx] = 5
	startCredits := gs.Player.Credits

	result := market.Sell(gs, goodIdx, 3)
	if !result.Success {
		t.Fatalf("sell failed: %s", result.Message)
	}
	if gs.Player.Credits <= startCredits {
		t.Error("credits should have increased")
	}
	if gs.Player.Cargo[goodIdx] != 2 {
		t.Errorf("cargo: got %d, want 2", gs.Player.Cargo[goodIdx])
	}
}

func TestSellInsufficientGoods(t *testing.T) {
	gs := newTestGame(t)

	result := market.Sell(gs, 0, 1)
	if result.Success {
		t.Error("should fail with no goods")
	}
}
