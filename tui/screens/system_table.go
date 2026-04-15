package screens

import (
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"

	"github.com/the4ofus/spacetrader-tui/internal/formula"
	"github.com/the4ofus/spacetrader-tui/internal/game"
	"github.com/the4ofus/spacetrader-tui/internal/gamedata"
	"github.com/the4ofus/spacetrader-tui/internal/market"
)

type sortColumn int

const (
	colName sortColumn = iota
	colDist
	colTech
	colGov
	colResource
)

type sortDir int

const (
	sortAsc sortDir = iota
	sortDesc
)

type systemEntry struct {
	sysIdx       int
	name         string
	dist         float64
	techLvl      gamedata.TechLevel
	techStr      string
	govStr       string
	resource     gamedata.Resource
	resStr       string
	sizeStr      string
	hasWormhole  bool
	visited      bool
	isCurrent    bool
	bookmarked   bool
	bookmarkNote string
	bookmarkDay  int
}

func buildSystemEntries(gs *game.GameState, indices []int) []systemEntry {
	refreshBookmarks(gs)
	cur := gs.Data.Systems[gs.CurrentSystemID]
	entries := make([]systemEntry, len(indices))
	for i, idx := range indices {
		sys := gs.Data.Systems[idx]
		bm, hasBM := gs.GetBookmark(idx)
		entries[i] = systemEntry{
			sysIdx:       idx,
			name:         sys.Name,
			dist:         formula.Distance(cur.X, cur.Y, sys.X, sys.Y),
			techLvl:      sys.TechLevel,
			techStr:      shortTech(sys.TechLevel),
			govStr:       sys.PoliticalSystem.String(),
			resource:     sys.Resource,
			resStr:       shortResource(sys.Resource),
			sizeStr:      sys.Size.String(),
			hasWormhole:  game.IsWormholeSystem(gs, idx),
			visited:      gs.Systems[idx].Visited,
			isCurrent:    idx == gs.CurrentSystemID,
			bookmarked:   hasBM,
			bookmarkNote: bm.Note,
			bookmarkDay:  bm.Day,
		}
	}
	return entries
}

func buildAllSystemEntries(gs *game.GameState) []systemEntry {
	indices := make([]int, len(gs.Data.Systems))
	for i := range indices {
		indices[i] = i
	}
	return buildSystemEntries(gs, indices)
}

func applyFilterAndSort(entries []systemEntry, filterText string, col sortColumn, dir sortDir) []systemEntry {
	result := filterSystemEntries(entries, filterText)
	sortSystemEntries(result, col, dir)
	return result
}

func filterSystemEntries(entries []systemEntry, text string) []systemEntry {
	if text == "" {
		result := make([]systemEntry, len(entries))
		copy(result, entries)
		return result
	}
	f := strings.ToLower(text)
	var result []systemEntry
	for _, e := range entries {
		if f == "!" || f == "bookmarked" {
			if e.bookmarked {
				result = append(result, e)
			}
			continue
		}
		if f == "*" || f == "visited" {
			if e.visited {
				result = append(result, e)
			}
			continue
		}
		if strings.Contains(strings.ToLower(e.name), f) ||
			strings.Contains(strings.ToLower(e.techStr), f) ||
			strings.Contains(strings.ToLower(e.govStr), f) ||
			strings.Contains(strings.ToLower(e.resStr), f) ||
			strings.Contains(strings.ToLower(e.bookmarkNote), f) {
			result = append(result, e)
		}
	}
	return result
}

func sortSystemEntries(entries []systemEntry, col sortColumn, dir sortDir) {
	sort.SliceStable(entries, func(i, j int) bool {
		var less bool
		switch col {
		case colName:
			less = strings.ToLower(entries[i].name) < strings.ToLower(entries[j].name)
		case colDist:
			less = entries[i].dist < entries[j].dist
		case colTech:
			less = entries[i].techLvl < entries[j].techLvl
		case colGov:
			less = strings.ToLower(entries[i].govStr) < strings.ToLower(entries[j].govStr)
		case colResource:
			less = strings.ToLower(entries[i].resStr) < strings.ToLower(entries[j].resStr)
		}
		if dir == sortDesc {
			return !less
		}
		return less
	})
}

func newFilterInput() textinput.Model {
	ti := textinput.New()
	ti.Placeholder = "type to filter..."
	ti.CharLimit = 30
	return ti
}

func sortedHeader(label string, col, activeCol sortColumn, dir sortDir) string {
	if col == activeCol {
		if dir == sortAsc {
			return label + "^"
		}
		return label + "v"
	}
	return label
}

func shortTech(t gamedata.TechLevel) string {
	switch t {
	case gamedata.TechPreAgricultural:
		return "Pre-ag"
	case gamedata.TechAgricultural:
		return "Agri"
	case gamedata.TechMedieval:
		return "Medieval"
	case gamedata.TechRenaissance:
		return "Renais"
	case gamedata.TechEarlyIndustrial:
		return "Early Ind"
	case gamedata.TechIndustrial:
		return "Industrial"
	case gamedata.TechPostIndustrial:
		return "Post-ind"
	case gamedata.TechHiTech:
		return "Hi-tech"
	}
	return t.String()
}

func shortResource(r gamedata.Resource) string {
	switch r {
	case gamedata.ResourceNone:
		return ""
	case gamedata.ResourceMineralRich:
		return "+Ore cheap"
	case gamedata.ResourceMineralPoor:
		return "-Ore pricey"
	case gamedata.ResourceDesert:
		return "-Water pricey"
	case gamedata.ResourceSweetOceans:
		return "+Water cheap"
	case gamedata.ResourceRichSoil:
		return "+Food cheap"
	case gamedata.ResourcePoorSoil:
		return "-Food pricey"
	case gamedata.ResourceRichFauna:
		return "+Furs cheap"
	case gamedata.ResourceLifeless:
		return "-Furs pricey"
	case gamedata.ResourceWeirdMushrooms:
		return "+Narco cheap"
	case gamedata.ResourceSpecialHerbs:
		return "+Meds cheap"
	case gamedata.ResourceArtistic:
		return "+Games cheap"
	case gamedata.ResourceWarlike:
		return "+Arms cheap"
	}
	return r.String()
}

func eventOrNone(event string) string {
	if event == "" {
		return "none"
	}
	return event
}

func colorResource(r gamedata.Resource, text string) string {
	switch r {
	case gamedata.ResourceNone:
		return text
	case gamedata.ResourceMineralRich, gamedata.ResourceSweetOceans, gamedata.ResourceRichFauna,
		gamedata.ResourceRichSoil, gamedata.ResourceWeirdMushrooms, gamedata.ResourceSpecialHerbs,
		gamedata.ResourceArtistic, gamedata.ResourceWarlike:
		return SuccessStyle.Render(text)
	case gamedata.ResourceMineralPoor, gamedata.ResourceDesert, gamedata.ResourceLifeless,
		gamedata.ResourcePoorSoil:
		return DangerStyle.Render(text)
	}
	return text
}

func resourceTradeHints(goods []gamedata.GoodDef, r gamedata.Resource) (cheap, expensive string) {
	resName := r.String()
	var cheapGoods, expensiveGoods []string
	for _, g := range goods {
		if g.CheapResource == resName {
			cheapGoods = append(cheapGoods, g.Name)
		}
		if g.ExpensiveResource == resName {
			expensiveGoods = append(expensiveGoods, g.Name)
		}
	}
	return strings.Join(cheapGoods, ", "), strings.Join(expensiveGoods, ", ")
}

func refreshBookmarks(gs *game.GameState) {
	for _, bm := range gs.Bookmarks {
		newNote := autoBookmarkNote(gs, bm.SystemIdx)
		if newNote != bm.Note {
			gs.UpdateBookmark(bm.SystemIdx, newNote)
		}
	}
}

var goodLetters = [game.NumGoods]string{"W", "U", "F", "O", "G", "A", "D", "C", "N", "R"}

func renderGoodsColumn(gs *game.GameState, sysIdx int) string {
	sysState := gs.Systems[sysIdx]
	info, hasInfo := gs.GetTradeInfo(sysIdx)
	stale := false
	if hasInfo {
		stale, _ = gs.IsTradeInfoStale(sysIdx)
	}

	var result string
	for g, good := range gs.Data.Goods {
		available := sysState.Prices[g] > 0
		if !available {
			result += DimStyle.Render(goodLetters[g])
			continue
		}
		if !hasInfo || stale {
			result += goodLetters[g]
			continue
		}
		avg := market.AveragePrice(good, gs.Data.Systems)
		price := info.Prices[g]
		if avg == 0 {
			result += goodLetters[g]
			continue
		}
		pct := (price - avg) * 100 / avg
		if pct <= -5 {
			result += SuccessStyle.Render(goodLetters[g])
		} else if pct >= 5 {
			result += DangerStyle.Render(goodLetters[g])
		} else {
			result += goodLetters[g]
		}
	}
	return result
}

func autoBookmarkNote(gs *game.GameState, sysIdx int) string {
	sys := gs.Data.Systems[sysIdx]
	sysState := gs.Systems[sysIdx]
	var parts []string
	if sysState.Event != "" {
		parts = append(parts, sysState.Event)
	}
	res := shortResource(sys.Resource)
	if res != "" {
		parts = append(parts, res)
	}
	parts = append(parts, shortTech(sys.TechLevel))
	return strings.Join(parts, ", ")
}
