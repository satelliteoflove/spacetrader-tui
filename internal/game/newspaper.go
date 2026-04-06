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

func GenerateEvents(gs *GameState) {
	for i := range gs.Systems {
		changed := false
		if gs.Rand.Intn(100) < 5 {
			gs.Systems[i].Event = eventNames[gs.Rand.Intn(len(eventNames))]
			changed = true
		} else if gs.Systems[i].Event != "" && gs.Rand.Intn(100) < 30 {
			gs.Systems[i].Event = ""
			changed = true
		}
		if changed {
			RefreshSystemPrices(gs, i)
		}
	}
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
			if dist < 20 {
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
