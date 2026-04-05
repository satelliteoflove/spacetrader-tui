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
	sysIdx    int
	name      string
	dist      float64
	techLvl   gamedata.TechLevel
	techStr   string
	govStr    string
	resStr    string
	visited   bool
	isCurrent bool
}

func buildSystemEntries(gs *game.GameState, indices []int) []systemEntry {
	cur := gs.Data.Systems[gs.CurrentSystemID]
	entries := make([]systemEntry, len(indices))
	for i, idx := range indices {
		sys := gs.Data.Systems[idx]
		entries[i] = systemEntry{
			sysIdx:    idx,
			name:      sys.Name,
			dist:      formula.Distance(cur.X, cur.Y, sys.X, sys.Y),
			techLvl:   sys.TechLevel,
			techStr:   shortTech(sys.TechLevel),
			govStr:    sys.PoliticalSystem.String(),
			resStr:    shortResource(sys.Resource),
			visited:   gs.Systems[idx].Visited,
			isCurrent: idx == gs.CurrentSystemID,
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
			strings.Contains(strings.ToLower(e.resStr), f) {
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
