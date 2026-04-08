package game

import (
	"fmt"
	"math/rand"

	"github.com/the4ofus/spacetrader-tui/internal/formula"
	"github.com/the4ofus/spacetrader-tui/internal/gamedata"
)

var eventNames = []string{
	"Drought", "Cold", "Crop Failure", "War", "Boredom",
	"Plague", "Lack of Workers", "Artistic",
}

const minEventDays = 3

func GenerateEvents(gs *GameState) {
	for i := range gs.Systems {
		changed := false
		if gs.Systems[i].Event == "" && gs.Rand.Intn(100) < 5 {
			event := eventNames[gs.Rand.Intn(len(eventNames))]
			if eventHasTradeableGoods(gs, i, event) {
				gs.Systems[i].Event = event
				gs.Systems[i].EventDay = gs.Day
				changed = true
			}
		} else if gs.Systems[i].Event != "" && gs.Day-gs.Systems[i].EventDay >= minEventDays && gs.Rand.Intn(100) < 30 {
			gs.Systems[i].Event = ""
			gs.Systems[i].EventDay = 0
			changed = true
			removeNewsForSystem(gs, i)
		}
		if changed {
			RefreshSystemPrices(gs, i)
		}
	}
}

func eventHasTradeableGoods(gs *GameState, sysIdx int, event string) bool {
	sys := gs.Data.Systems[sysIdx]
	for _, good := range gs.Data.Goods {
		if good.PriceIncreaseEvent == event || good.PriceDecreaseEvent == event {
			if good.MinTech <= sys.TechLevel && sys.TechLevel <= good.MaxTech {
				return true
			}
		}
	}
	return false
}

func SystemMasthead(gs *GameState) string {
	sys := gs.Data.Systems[gs.CurrentSystemID]
	polData := gamedata.PoliticalSystems[sys.PoliticalSystem]
	idx := gs.CurrentSystemID % 3
	return polData.NewspaperNames[idx]
}

func GenerateNewspaper(gs *GameState) []string {
	sys := gs.Data.Systems[gs.CurrentSystemID]
	sysState := gs.Systems[gs.CurrentSystemID]
	rng := gs.Rand

	var headlines []string

	if sysState.Event != "" {
		h := eventHeadline(sysState.Event, sys.Name)
		headlines = append(headlines, h)
		addNewsEntry(gs, h, sys.Name, gs.CurrentSystemID)
	}

	for i, neighbor := range gs.Data.Systems {
		if i == gs.CurrentSystemID {
			continue
		}
		if gs.Systems[i].Event != "" {
			dist := formula.Distance(sys.X, sys.Y, neighbor.X, neighbor.Y)
			if dist < 30 {
				h := eventHeadline(gs.Systems[i].Event, neighbor.Name)
				headlines = append(headlines, h)
				addNewsEntry(gs, h, neighbor.Name, i)
			}
		}
	}

	headlines = append(headlines, flavorHeadline(sys, rng))

	if len(headlines) > 5 {
		headlines = headlines[:5]
	}

	return headlines
}

func addNewsEntry(gs *GameState, headline, system string, sysIdx int) {
	for _, e := range gs.NewsLog {
		if e.Headline == headline && e.Day == gs.Day {
			return
		}
	}
	gs.NewsLog = append(gs.NewsLog, NewsEntry{
		Headline:  headline,
		System:    system,
		SystemIdx: sysIdx,
		Day:       gs.Day,
	})
	if len(gs.NewsLog) > 50 {
		gs.NewsLog = gs.NewsLog[len(gs.NewsLog)-50:]
	}
}

func removeNewsForSystem(gs *GameState, sysIdx int) {
	filtered := gs.NewsLog[:0]
	for _, e := range gs.NewsLog {
		if e.SystemIdx != sysIdx {
			filtered = append(filtered, e)
		}
	}
	gs.NewsLog = filtered
}

func eventHeadline(event string, systemName string) string {
	switch event {
	case "Drought":
		return fmt.Sprintf("Water shortage on %s!", systemName)
	case "Cold":
		return fmt.Sprintf("Freezing temperatures reported on %s.", systemName)
	case "Crop Failure":
		return fmt.Sprintf("Crop failure devastates %s farmers.", systemName)
	case "War":
		return fmt.Sprintf("Conflict erupts on %s -- arms dealers take note.", systemName)
	case "Boredom":
		return fmt.Sprintf("Citizens of %s complain of boredom.", systemName)
	case "Plague":
		return fmt.Sprintf("Plague sweeps through %s! Medicine desperately needed.", systemName)
	case "Lack of Workers":
		return fmt.Sprintf("Labor shortage on %s drives up machine prices.", systemName)
	case "Artistic":
		return fmt.Sprintf("Cultural renaissance on %s -- demand for games soars.", systemName)
	}
	return fmt.Sprintf("Unusual activity reported on %s.", systemName)
}

type NewsBriefing struct {
	SystemName      string
	Distance        float64
	InRange         bool
	EventActive     bool
	Blurb           string
	PriceLines      []string
	CargoAlerts     []string
	SecurityWarning string
}

func GenerateNewsBriefing(gs *GameState, entry NewsEntry) NewsBriefing {
	sysIdx := entry.SystemIdx
	sys := gs.Data.Systems[sysIdx]
	cur := gs.Data.Systems[gs.CurrentSystemID]
	dist := formula.Distance(cur.X, cur.Y, sys.X, sys.Y)
	fuelRange := float64(gs.Player.Ship.Fuel)

	brief := NewsBriefing{
		SystemName: sys.Name,
		Distance:   dist,
		InRange:    dist <= fuelRange,
	}

	event := gs.Systems[sysIdx].Event
	if event == "" {
		brief.EventActive = false
		brief.Blurb = fmt.Sprintf("Reports of unusual activity on %s appear to have subsided. Prices have normalized.", sys.Name)
		return brief
	}
	brief.EventActive = true
	brief.Blurb = eventBlurb(event, sys.Name)

	curPrices := gs.Systems[gs.CurrentSystemID].Prices
	destPrices := gs.Systems[sysIdx].Prices

	for g, good := range gs.Data.Goods {
		if destPrices[g] < 0 {
			continue
		}

		if good.PriceIncreaseEvent == event {
			if curPrices[g] > 0 {
				profit := destPrices[g] - curPrices[g]
				brief.PriceLines = append(brief.PriceLines,
					fmt.Sprintf("%-10s here %d cr | there ~%d cr (est. %+d/unit)",
						good.Name+":", curPrices[g], destPrices[g], profit))
			} else {
				brief.PriceLines = append(brief.PriceLines,
					fmt.Sprintf("%-10s there ~%d cr (not sold here)",
						good.Name+":", destPrices[g]))
			}
		}

		if good.PriceDecreaseEvent == event {
			if curPrices[g] > 0 {
				profit := destPrices[g] - curPrices[g]
				brief.PriceLines = append(brief.PriceLines,
					fmt.Sprintf("%-10s here %d cr | there ~%d cr (est. %+d/unit)",
						good.Name+":", curPrices[g], destPrices[g], profit))
			} else {
				brief.PriceLines = append(brief.PriceLines,
					fmt.Sprintf("%-10s there ~%d cr (not sold here)",
						good.Name+":", destPrices[g]))
			}
		}

		if (good.PriceIncreaseEvent == event || good.PriceDecreaseEvent == event) && gs.Player.Cargo[g] > 0 {
			brief.CargoAlerts = append(brief.CargoAlerts,
				fmt.Sprintf("You have %d %s in cargo!", gs.Player.Cargo[g], good.Name))
		}
	}

	strictGov := sys.PoliticalSystem == gamedata.PolDictatorship ||
		sys.PoliticalSystem == gamedata.PolFascist ||
		sys.PoliticalSystem == gamedata.PolMilitary
	if strictGov {
		hasIllegal := false
		for g, good := range gs.Data.Goods {
			if !good.Legal && gs.Player.Cargo[g] > 0 {
				hasIllegal = true
				break
			}
		}
		if hasIllegal {
			brief.SecurityWarning = "Strict government -- illegal goods in your cargo will draw attention!"
		} else {
			brief.SecurityWarning = "Strict government -- avoid carrying illegal goods here."
		}
	}

	return brief
}

func eventBlurb(event, name string) string {
	switch event {
	case "Drought":
		return fmt.Sprintf("Severe drought conditions across %s have depleted water reserves to critical levels. Local traders report Water demand up 50%% as authorities scramble to secure emergency supplies.", name)
	case "Cold":
		return fmt.Sprintf("A brutal cold snap has gripped %s, driving demand for Furs up 50%% as residents struggle to stay warm. Merchants with surplus furs stand to profit handsomely.", name)
	case "Crop Failure":
		return fmt.Sprintf("Devastating crop failures across %s's agricultural districts have sent food prices soaring. Food demand is up 50%% as reserves run dangerously low.", name)
	case "War":
		return fmt.Sprintf("Open conflict has erupted across %s as rival factions battle for control. Arms dealers are flooding the region -- Ore and Firearms demand is up 50%% as both sides stockpile.", name)
	case "Boredom":
		return fmt.Sprintf("Citizens of %s are suffering from widespread ennui. Demand for Games is up 50%%, while Narcotics traders may also find willing buyers in the disaffected population.", name)
	case "Plague":
		return fmt.Sprintf("A deadly plague is sweeping through %s, overwhelming local medical facilities. Medicine demand is up 50%% as the death toll climbs. Any trader carrying medical supplies would be welcomed.", name)
	case "Lack of Workers":
		return fmt.Sprintf("A critical labor shortage on %s has crippled local industry. Demand for Machines and Robots is up 50%% as factories try to automate their way through the crisis.", name)
	case "Artistic":
		return fmt.Sprintf("A cultural renaissance is flourishing on %s! Demand for Games has surged 50%% as the population embraces new forms of entertainment and artistic expression.", name)
	}
	return fmt.Sprintf("Unusual activity has been reported on %s. Market conditions may be affected.", name)
}

func flavorHeadline(sys gamedata.SystemDef, rng *rand.Rand) string {
	flavors := []string{
		fmt.Sprintf("Trade volumes steady at %s spaceport.", sys.Name),
		fmt.Sprintf("%s authorities remind visitors to declare all cargo.", sys.Name),
		"Fuel prices remain stable across the sector.",
		"Galactic merchant guild reports record trade quarter.",
		fmt.Sprintf("New governor appointed on %s.", sys.Name),
	}
	return flavors[rng.Intn(len(flavors))]
}
