package game

import (
	"math"
	"math/rand"

	"github.com/the4ofus/spacetrader-tui/internal/formula"
)

var mercenaryNames = []string{
	"Alyssa", "Armatur", "Bentos", "C2U2", "Chi'Ti",
	"Crystal", "Dane", "Deirdre", "Doc", "Draco",
	"Iranda", "Jeremiah", "Jujubal", "Krydon", "Luis",
	"Mercedez", "Milete", "Muri-L", "Mystyc", "Nandi",
	"Orestes", "Pancho", "PS37", "Quarck", "Sosumi",
	"Uma", "Wesley", "Wonton", "Yorvick",
}

func randomSkill(rng *rand.Rand) int {
	return 1 + rng.Intn(5) + rng.Intn(6)
}

func GenerateMercenaries(rng *rand.Rand, numSystems int, startIdx int, systems [][2]int, maxRange int) []Mercenary {
	mercs := make([]Mercenary, len(mercenaryNames))
	sysCounts := make([]int, numSystems)

	for i, name := range mercenaryNames {
		skills := [formula.NumSkills]int{
			randomSkill(rng),
			randomSkill(rng),
			randomSkill(rng),
			randomSkill(rng),
		}

		sysIdx := pickMercSystem(rng, startIdx, systems, maxRange, sysCounts)
		mercs[i] = Mercenary{
			Name:      name,
			Skills:    skills,
			SystemIdx: sysIdx,
		}
		if sysIdx >= 0 {
			sysCounts[sysIdx]++
		}
	}
	return mercs
}

func pickMercSystem(rng *rand.Rand, startIdx int, systems [][2]int, maxRange int, counts []int) int {
	start := systems[startIdx]
	var candidates []int
	for i, sys := range systems {
		if i == startIdx {
			continue
		}
		dx := float64(start[0] - sys[0])
		dy := float64(start[1] - sys[1])
		dist := math.Sqrt(dx*dx + dy*dy)
		if int(math.Ceil(dist)) <= maxRange*2 && counts[i] < 3 {
			candidates = append(candidates, i)
		}
	}
	if len(candidates) == 0 {
		for i := range systems {
			if i != startIdx && counts[i] < 3 {
				candidates = append(candidates, i)
			}
		}
	}
	if len(candidates) == 0 {
		return rng.Intn(len(systems))
	}
	return candidates[rng.Intn(len(candidates))]
}

func AvailableMercenaries(gs *GameState) []int {
	var indices []int
	hired := map[string]bool{}
	for _, m := range gs.Player.Crew {
		hired[m.Name] = true
	}
	for i, m := range gs.Mercenaries {
		if m.SystemIdx == gs.CurrentSystemID && !hired[m.Name] {
			indices = append(indices, i)
		}
	}
	return indices
}

func FreeCrewQuarters(gs *GameState) int {
	shipDef := gs.Data.Ships[gs.Player.Ship.TypeID]
	return shipDef.CrewQuarters - 1 - len(gs.Player.Crew)
}

func HireMercenary(gs *GameState, mercIdx int) (bool, string) {
	if mercIdx < 0 || mercIdx >= len(gs.Mercenaries) {
		return false, "Invalid mercenary."
	}
	merc := &gs.Mercenaries[mercIdx]
	if merc.SystemIdx != gs.CurrentSystemID {
		return false, "Mercenary is not at this system."
	}

	if FreeCrewQuarters(gs) <= 0 {
		return false, "No crew quarters available."
	}

	merc.SystemIdx = -1
	gs.Player.Crew = append(gs.Player.Crew, *merc)
	return true, merc.Name + " hired."
}

func FireMercenary(gs *GameState, idx int) (bool, string) {
	if idx < 0 || idx >= len(gs.Player.Crew) {
		return false, "Invalid crew member."
	}
	merc := gs.Player.Crew[idx]
	if merc.IsQuest {
		return false, merc.Name + " is a passenger and cannot be dismissed."
	}

	name := merc.Name
	gs.Player.Crew = append(gs.Player.Crew[:idx], gs.Player.Crew[idx+1:]...)

	for i, m := range gs.Mercenaries {
		if m.Name == name {
			gs.Mercenaries[i].SystemIdx = pickNearbySystem(gs)
			break
		}
	}

	return true, name + " fired."
}

func pickNearbySystem(gs *GameState) int {
	cur := gs.Data.Systems[gs.CurrentSystemID]
	maxRange := gs.EffectiveRange()
	var candidates []int
	for i, sys := range gs.Data.Systems {
		if i == gs.CurrentSystemID {
			continue
		}
		dist := formula.Distance(cur.X, cur.Y, sys.X, sys.Y)
		if int(math.Ceil(dist)) <= maxRange*2 {
			candidates = append(candidates, i)
		}
	}
	if len(candidates) == 0 {
		return gs.Rand.Intn(len(gs.Data.Systems))
	}
	return candidates[gs.Rand.Intn(len(candidates))]
}

func NthLowestSkill(skills [formula.NumSkills]int, n int) int {
	ids := []int{0, 1, 2, 3}
	for j := 0; j < 3; j++ {
		for i := 0; i < 3-j; i++ {
			if skills[ids[i]] > skills[ids[i+1]] {
				ids[i], ids[i+1] = ids[i+1], ids[i]
			}
		}
	}
	return ids[n-1]
}

func AddQuestCrew(gs *GameState, name string, skills [formula.NumSkills]int) bool {
	if FreeCrewQuarters(gs) <= 0 {
		return false
	}
	gs.Player.Crew = append(gs.Player.Crew, Mercenary{
		Name:      name,
		Skills:    skills,
		SystemIdx: -1,
		IsQuest:   true,
	})
	return true
}

func RemoveQuestCrew(gs *GameState, name string) {
	for i, m := range gs.Player.Crew {
		if m.Name == name && m.IsQuest {
			gs.Player.Crew = append(gs.Player.Crew[:i], gs.Player.Crew[i+1:]...)
			return
		}
	}
}

func HasQuestCrew(gs *GameState, name string) bool {
	for _, m := range gs.Player.Crew {
		if m.Name == name && m.IsQuest {
			return true
		}
	}
	return false
}

func ClearCrewAndResetQuests(gs *GameState) {
	for _, m := range gs.Player.Crew {
		if m.IsQuest {
			switch m.Name {
			case "Jarek":
				gs.SetQuestState(QuestJarek, QuestUnavailable)
				gs.SetQuestProgress(QuestJarek, 0)
			case "Wild":
				gs.SetQuestState(QuestWild, QuestUnavailable)
				gs.SetQuestProgress(QuestWild, 0)
			}
		} else {
			for i, pm := range gs.Mercenaries {
				if pm.Name == m.Name {
					gs.Mercenaries[i].SystemIdx = gs.CurrentSystemID
					break
				}
			}
		}
	}
	gs.Player.Crew = nil
}

func CreateZeethibal(gs *GameState) {
	lowest1 := NthLowestSkill(gs.Player.Skills, 1)
	lowest2 := NthLowestSkill(gs.Player.Skills, 2)

	var skills [formula.NumSkills]int
	for i := range skills {
		if i == lowest1 {
			skills[i] = 10
		} else if i == lowest2 {
			skills[i] = 8
		} else {
			skills[i] = 5
		}
	}

	kravat := findSystem(gs, "Kravat")
	if kravat < 0 {
		return
	}

	for _, m := range gs.Mercenaries {
		if m.Name == "Zeethibal" {
			return
		}
	}

	gs.Mercenaries = append(gs.Mercenaries, Mercenary{
		Name:      "Zeethibal",
		Skills:    skills,
		SystemIdx: kravat,
		IsQuest:   true,
	})
}
