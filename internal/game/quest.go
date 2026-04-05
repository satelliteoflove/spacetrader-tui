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
