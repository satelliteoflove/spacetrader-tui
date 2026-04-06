package market

import (
	"math/rand"

	"github.com/the4ofus/spacetrader-tui/internal/formula"
	"github.com/the4ofus/spacetrader-tui/internal/gamedata"
)

func CalculatePrice(good gamedata.GoodDef, sys gamedata.SystemDef, event string, traderSkill int, rng *rand.Rand) int {
	if good.MinTech > sys.TechLevel || sys.TechLevel > good.MaxTech {
		return -1
	}

	price := formula.BasePrice(good, sys, event, rng)

	if traderSkill > 0 {
		discount := traderSkill
		if discount > 10 {
			discount = 10
		}
		price = price * (100 - discount) / 100
	}

	if price < 1 {
		price = 1
	}
	return price
}

func AveragePrice(good gamedata.GoodDef, systems []gamedata.SystemDef) int {
	total := 0
	count := 0
	for _, sys := range systems {
		if good.MinTech > sys.TechLevel || sys.TechLevel > good.MaxTech {
			continue
		}
		total += formula.BasePrice(good, sys, "", nil)
		count++
	}
	if count == 0 {
		return good.BasePrice
	}
	return total / count
}

func PriceVsAverage(localPrice int, avgPrice int) string {
	if avgPrice == 0 {
		return ""
	}
	diff := localPrice - avgPrice
	pct := diff * 100 / avgPrice
	switch {
	case pct <= -15:
		return "very cheap"
	case pct <= -5:
		return "cheap"
	case pct >= 15:
		return "very expensive"
	case pct >= 5:
		return "expensive"
	default:
		return "average"
	}
}
