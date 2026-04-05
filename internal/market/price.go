package market

import (
	"math/rand"

	"github.com/the4ofus/spacetrader-tui/internal/gamedata"
)

func CalculatePrice(good gamedata.GoodDef, sys gamedata.SystemDef, event string, traderSkill int, rng *rand.Rand) int {
	if good.MinTech > sys.TechLevel || sys.TechLevel > good.MaxTech {
		return -1
	}

	price := good.BasePrice

	techRange := int(good.MaxTech - good.MinTech)
	if techRange > 0 {
		techPos := int(sys.TechLevel - good.MinTech)
		price = price + (price * techPos / techRange / 2) - (price / 4)
	}

	if good.ExpensiveResource != "" && sys.Resource.String() == good.ExpensiveResource {
		price = price * 150 / 100
	}
	if good.CheapResource != "" && sys.Resource.String() == good.CheapResource {
		price = price * 70 / 100
	}

	if (sys.PoliticalSystem == gamedata.PolDictatorship ||
		sys.PoliticalSystem == gamedata.PolFascist ||
		sys.PoliticalSystem == gamedata.PolMilitary) && !good.Legal {
		price = price * 150 / 100
	}

	if event != "" {
		if event == good.PriceIncreaseEvent {
			price = price * 150 / 100
		}
		if event == good.PriceDecreaseEvent {
			price = price * 70 / 100
		}
	}

	variance := good.Variance
	if variance > 0 && rng != nil {
		delta := rng.Intn(2*variance+1) - variance
		price += delta
	}

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
		price := good.BasePrice
		techRange := int(good.MaxTech - good.MinTech)
		if techRange > 0 {
			techPos := int(sys.TechLevel - good.MinTech)
			price = price + (price * techPos / techRange / 2) - (price / 4)
		}
		if good.ExpensiveResource != "" && sys.Resource.String() == good.ExpensiveResource {
			price = price * 150 / 100
		}
		if good.CheapResource != "" && sys.Resource.String() == good.CheapResource {
			price = price * 70 / 100
		}
		total += price
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
