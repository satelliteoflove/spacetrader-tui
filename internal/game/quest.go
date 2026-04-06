package game

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
	NumQuests
)

type QuestState int

const (
	QuestUnavailable QuestState = iota
	QuestAvailable
	QuestActive
	QuestComplete
)

type QuestData struct {
	States     [NumQuests]QuestState `json:"states"`
	Progress   [NumQuests]int        `json:"progress"`
	TribbleQty int                   `json:"tribble_qty"`
}

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

	default:
		return QuestUrgencyFresh
	}
}
