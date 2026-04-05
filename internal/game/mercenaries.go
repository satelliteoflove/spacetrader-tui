package game

import "github.com/the4ofus/spacetrader-tui/internal/formula"

var MercenaryPool = []Mercenary{
	{Name: "Alyssa", Skills: [formula.NumSkills]int{7, 3, 2, 5}, Wage: 30},
	{Name: "Bones", Skills: [formula.NumSkills]int{2, 2, 4, 8}, Wage: 35},
	{Name: "Cassie", Skills: [formula.NumSkills]int{5, 6, 3, 3}, Wage: 25},
	{Name: "Drake", Skills: [formula.NumSkills]int{3, 8, 2, 2}, Wage: 40},
	{Name: "Echo", Skills: [formula.NumSkills]int{8, 3, 4, 2}, Wage: 35},
	{Name: "Faye", Skills: [formula.NumSkills]int{2, 4, 8, 3}, Wage: 30},
	{Name: "Grit", Skills: [formula.NumSkills]int{4, 7, 2, 5}, Wage: 35},
	{Name: "Harper", Skills: [formula.NumSkills]int{3, 2, 7, 6}, Wage: 30},
	{Name: "Iris", Skills: [formula.NumSkills]int{6, 5, 5, 3}, Wage: 30},
	{Name: "Jet", Skills: [formula.NumSkills]int{4, 3, 3, 7}, Wage: 25},
	{Name: "Knox", Skills: [formula.NumSkills]int{2, 9, 1, 3}, Wage: 45},
	{Name: "Luna", Skills: [formula.NumSkills]int{9, 2, 3, 4}, Wage: 45},
	{Name: "Mason", Skills: [formula.NumSkills]int{3, 5, 6, 4}, Wage: 30},
	{Name: "Nyx", Skills: [formula.NumSkills]int{5, 4, 4, 6}, Wage: 30},
	{Name: "Orla", Skills: [formula.NumSkills]int{4, 6, 5, 3}, Wage: 30},
}

func AvailableMercenaries(gs *GameState) []Mercenary {
	hired := map[string]bool{}
	for _, m := range gs.Player.Crew {
		hired[m.Name] = true
	}

	count := 2 + gs.Rand.Intn(3)

	var available []Mercenary
	perm := gs.Rand.Perm(len(MercenaryPool))
	for _, idx := range perm {
		m := MercenaryPool[idx]
		if !hired[m.Name] {
			available = append(available, m)
			if len(available) >= count {
				break
			}
		}
	}
	return available
}

func HireMercenary(gs *GameState, merc Mercenary) (bool, string) {
	shipDef := gs.Data.Ships[gs.Player.Ship.TypeID]
	if len(gs.Player.Crew)+1 > shipDef.CrewQuarters-1 {
		return false, "No crew quarters available."
	}
	if gs.Player.Credits < merc.Wage {
		return false, "Not enough credits for signing bonus."
	}
	gs.Player.Credits -= merc.Wage
	gs.Player.Crew = append(gs.Player.Crew, merc)
	return true, merc.Name + " hired."
}

func FireMercenary(gs *GameState, idx int) (bool, string) {
	if idx < 0 || idx >= len(gs.Player.Crew) {
		return false, "Invalid crew member."
	}
	name := gs.Player.Crew[idx].Name
	gs.Player.Crew = append(gs.Player.Crew[:idx], gs.Player.Crew[idx+1:]...)
	return true, name + " fired."
}
