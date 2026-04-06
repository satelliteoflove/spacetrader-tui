package formula

import (
	"math/rand"

	"github.com/the4ofus/spacetrader-tui/internal/gamedata"
)

func BasePrice(good gamedata.GoodDef, sys gamedata.SystemDef, event string, rng *rand.Rand) int {
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

	if rng != nil {
		variance := good.Variance
		if variance > 0 {
			delta := rng.Intn(2*variance+1) - variance
			price += delta
		}
	}

	if price < 1 {
		price = 1
	}
	return price
}
