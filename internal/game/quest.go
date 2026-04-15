package game

import "fmt"

type QuestID int

const (
	QuestNone QuestID = iota
	QuestDragonfly
	QuestSpaceMonster
	QuestScarab
	QuestAlienArtifact
	QuestJarek
	QuestJapori
	QuestGemulon
	QuestFehler
	QuestWild
	QuestReactor
	QuestTribbles
	QuestSkillIncrease
	QuestEraseRecord
	QuestCargoForSale
	QuestLotteryWinner
	QuestMoonForSale
	QuestFabricRip
	NumQuests
)

type QuestState int

const (
	QuestUnavailable QuestState = iota
	QuestAvailable
	QuestActive
	QuestComplete
)

type PendingReward struct {
	QuestID   QuestID `json:"quest_id"`
	Equipment string  `json:"equipment"`
	SystemIdx int     `json:"system_idx"`
}

type QuestData struct {
	States         [NumQuests]QuestState `json:"states"`
	Progress       [NumQuests]int        `json:"progress"`
	TribbleQty     int                   `json:"tribble_qty"`
	MonsterHull    int                   `json:"monster_hull"`
	DragonflyHull  int                   `json:"dragonfly_hull"`
	FabricRipDays  int                   `json:"fabric_rip_days"`
	HasSingularity bool                  `json:"has_singularity"`
	PendingRewards []PendingReward       `json:"pending_rewards"`
}

const (
	ReactorStatusNotStarted = 0
	ReactorStatusFuelOk     = 1
	ReactorStatusDate       = 20
	ReactorStatusDelivered  = 21
	ReactorStatusDone       = 22
	ReactorBays             = 5
	ReactorFuelBays         = 10
	ReactorTotalBays        = 15
)

const MonsterMaxHull = 500
const DragonflyMaxHull = 150

func (gs *GameState) QuestState(id QuestID) QuestState {
	return gs.Quests.States[id]
}

func (gs *GameState) SetQuestState(id QuestID, state QuestState) {
	gs.Quests.States[id] = state
}

func (gs *GameState) QuestProgress(id QuestID) int {
	return gs.Quests.Progress[id]
}

func (gs *GameState) SetQuestProgress(id QuestID, progress int) {
	gs.Quests.Progress[id] = progress
}

func (gs *GameState) QuestDescription(id QuestID) string {
	state := gs.Quests.States[id]
	progress := gs.Quests.Progress[id]

	switch id {
	case QuestDragonfly:
		path := []string{"Arouan", "Halley", "Regulus", "Linnet"}
		if progress >= len(path) {
			return "Complete"
		}
		start := 0
		if state == QuestActive {
			start = progress
		}
		var parts []string
		for i, name := range path {
			if i < start {
				parts = append(parts, DimMark+name+DimMark)
			} else if i == start {
				parts = append(parts, NextMark+name+NextMark)
			} else {
				parts = append(parts, name)
			}
		}
		return "Route: " + joinArrow(parts)

	case QuestJarek:
		if state == QuestActive {
			remaining := 10 - progress
			return fmt.Sprintf("Transport Jarek to Devidia (%d stops remaining)", remaining)
		}
		return "Transport Jarek to Devidia"

	case QuestGemulon:
		if state == QuestAvailable {
			remaining := 7 - (gs.Day - progress)
			return fmt.Sprintf("Warn Gemulon (%d days remaining)", remaining)
		}
		return "Warn Gemulon"

	case QuestFehler:
		if state == QuestAvailable {
			remaining := 5 - (gs.Day - progress)
			return fmt.Sprintf("Stop experiment at Deneb (%d days remaining)", remaining)
		}
		return "Stop experiment at Deneb"

	case QuestReactor:
		status := gs.Quests.Progress[id]
		if status == ReactorStatusDelivered {
			return "Collect Morgan's Laser at Nix"
		}
		fuelBays := ReactorFuelBays - (status-1)/2
		if fuelBays < 0 {
			fuelBays = 0
		}
		return fmt.Sprintf("Deliver reactor to Nix (%d bays, %d fuel bays)", ReactorBays, fuelBays)

	case QuestWild:
		return "Smuggle Wild to Kravat (police danger!)"

	case QuestJapori:
		carried := gs.Player.Cargo[findGoodIndex(gs, "Medicine")]
		return fmt.Sprintf("Deliver medicine to Japori (%d/10 carried)", carried)

	case QuestSpaceMonster:
		return "Destroy at Acamar"
	case QuestScarab:
		return "Find near a wormhole exit"
	case QuestAlienArtifact:
		return "Deliver to a Hi-tech system"
	case QuestMoonForSale:
		needed := 500000 - gs.Player.Credits
		if gs.Player.LoanBalance > 0 {
			return fmt.Sprintf("Need %d cr and pay off debt (%d cr)", needed, gs.Player.LoanBalance)
		}
		if needed > 0 {
			return fmt.Sprintf("Need %d more credits (500,000 to buy)", needed)
		}
		return "You can afford it! Look for it on the system menu"
	default:
		return ""
	}
}

const (
	DimMark  = "\x00dim\x00"
	NextMark = "\x00next\x00"
)

func joinArrow(parts []string) string {
	result := parts[0]
	for _, p := range parts[1:] {
		result += " -> " + p
	}
	return result
}

func (gs *GameState) QuestDestination(id QuestID) int {
	switch id {
	case QuestDragonfly:
		path := []string{"Arouan", "Halley", "Regulus", "Linnet"}
		progress := gs.Quests.Progress[id]
		if progress < len(path) {
			return findSystem(gs, path[progress])
		}
		return findSystem(gs, path[len(path)-1])
	case QuestJarek:
		return findSystem(gs, "Devidia")
	case QuestGemulon:
		return findSystem(gs, "Gemulon")
	case QuestFehler:
		return findSystem(gs, "Deneb")
	case QuestReactor:
		if gs.Quests.Progress[id] == ReactorStatusDelivered {
			return findSystem(gs, "Nix")
		}
		return findSystem(gs, "Nix")
	case QuestWild:
		return findSystem(gs, "Kravat")
	case QuestJapori:
		return findSystem(gs, "Japori")
	case QuestSpaceMonster:
		return findSystem(gs, "Acamar")
	case QuestAlienArtifact:
		return -1
	case QuestScarab:
		return -1
	case QuestMoonForSale:
		return -1
	}
	return -1
}

func findGoodIndex(gs *GameState, name string) int {
	for i, g := range gs.Data.Goods {
		if g.Name == name {
			return i
		}
	}
	return 0
}

type QuestUrgencyLevel int

const (
	QuestUrgencyNone   QuestUrgencyLevel = iota
	QuestUrgencyFresh
	QuestUrgencyStale
	QuestUrgencyCritical
)

func (gs *GameState) QuestUrgency() QuestUrgencyLevel {
	worst := QuestUrgencyNone

	for id := QuestID(1); id < NumQuests; id++ {
		state := gs.Quests.States[id]
		if state != QuestAvailable && state != QuestActive {
			continue
		}

		urgency := questUrgency(gs, id)
		if urgency > worst {
			worst = urgency
		}
	}
	return worst
}

func questUrgency(gs *GameState, id QuestID) QuestUrgencyLevel {
	switch id {
	case QuestGemulon:
		if gs.Quests.States[id] != QuestAvailable {
			return QuestUrgencyNone
		}
		remaining := 7 - (gs.Day - gs.Quests.Progress[id])
		if remaining <= 2 {
			return QuestUrgencyCritical
		}
		if remaining <= 4 {
			return QuestUrgencyStale
		}
		return QuestUrgencyFresh

	case QuestFehler:
		if gs.Quests.States[id] != QuestAvailable {
			return QuestUrgencyNone
		}
		remaining := 5 - (gs.Day - gs.Quests.Progress[id])
		if remaining <= 1 {
			return QuestUrgencyCritical
		}
		if remaining <= 3 {
			return QuestUrgencyStale
		}
		return QuestUrgencyFresh

	case QuestJarek:
		if gs.Quests.States[id] != QuestActive {
			return QuestUrgencyNone
		}
		elapsed := gs.Quests.Progress[id]
		if elapsed >= 8 {
			return QuestUrgencyCritical
		}
		if elapsed >= 5 {
			return QuestUrgencyStale
		}
		return QuestUrgencyFresh

	case QuestReactor:
		if gs.Quests.States[id] != QuestActive {
			return QuestUrgencyNone
		}
		status := gs.Quests.Progress[id]
		if status >= ReactorStatusDate-2 {
			return QuestUrgencyCritical
		}
		if status >= ReactorStatusDate-4 {
			return QuestUrgencyStale
		}
		return QuestUrgencyFresh

	default:
		return QuestUrgencyFresh
	}
}
