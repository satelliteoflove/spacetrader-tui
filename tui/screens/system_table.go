package screens

import (
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"

	"github.com/the4ofus/spacetrader-tui/internal/formula"
	"github.com/the4ofus/spacetrader-tui/internal/game"
	"github.com/the4ofus/spacetrader-tui/internal/gamedata"
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
	visited      bool
	isCurrent    bool
	bookmarked   bool
	bookmarkNote string
}

func buildSystemEntries(gs *game.GameState, indices []int) []systemEntry {
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
			visited:      gs.Systems[idx].Visited,
			isCurrent:    idx == gs.CurrentSystemID,
			bookmarked:   hasBM,
			bookmarkNote: bm.Note,
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
		return "+Minerals"
	case gamedata.ResourceWaterWorld:
		return "+Water"
	case gamedata.ResourceRichFauna:
		return "+Fauna"
	case gamedata.ResourceRichSoil:
		return "+Soil"
	case gamedata.ResourceGoodClinic:
		return "+Good med"
	case gamedata.ResourceRobotWorkers:
		return "+Robots"
	case gamedata.ResourceDesert:
		return "-Desert"
	case gamedata.ResourcePoor:
		return "-Poor"
	case gamedata.ResourceLifeless:
		return "-Lifeless"
	case gamedata.ResourcePoorSoil:
		return "-Poor soil"
	case gamedata.ResourcePoorClinic:
		return "-Poor med"
	case gamedata.ResourceLackOfWorkers:
		return "-Low labor"
	case gamedata.ResourceIndustrial:
		return "~Industrial"
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
	case gamedata.ResourceMineralRich, gamedata.ResourceWaterWorld, gamedata.ResourceRichFauna,
		gamedata.ResourceRichSoil, gamedata.ResourceGoodClinic, gamedata.ResourceRobotWorkers:
		return SuccessStyle.Render(text)
	case gamedata.ResourceDesert, gamedata.ResourcePoor, gamedata.ResourceLifeless,
		gamedata.ResourcePoorSoil, gamedata.ResourcePoorClinic, gamedata.ResourceLackOfWorkers:
		return DangerStyle.Render(text)
	case gamedata.ResourceIndustrial:
		return SelectedStyle.Render(text)
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
