package encounter

type EncounterType int

const (
	EncPolice EncounterType = iota
	EncPirate
	EncTrader
	EncFamousCaptain
	EncMarieCeleste
	EncBottle
)

func (e EncounterType) String() string {
	switch e {
	case EncPolice:
		return "Police"
	case EncPirate:
		return "Pirate"
	case EncTrader:
		return "Trader"
	case EncFamousCaptain:
		return "Famous Captain"
	case EncMarieCeleste:
		return "Marie Celeste"
	case EncBottle:
		return "Floating Bottle"
	}
	return "Unknown"
}

type Action int

const (
	ActionComply     Action = iota
	ActionBribe
	ActionFlee
	ActionFight
	ActionSurrender
	ActionTrade
	ActionDecline
)

func (a Action) String() string {
	switch a {
	case ActionComply:
		return "Comply"
	case ActionBribe:
		return "Bribe"
	case ActionFlee:
		return "Flee"
	case ActionFight:
		return "Fight"
	case ActionSurrender:
		return "Surrender"
	case ActionTrade:
		return "Trade"
	case ActionDecline:
		return "Decline"
	}
	return "Unknown"
}

type Encounter struct {
	Type    EncounterType
	Actions []Action
	Message string
}

type Outcome struct {
	Message       string
	CreditsChange int
	HullDamage    int
	CargoLost     map[int]int
	RecordChange  int
	RepChange     int
	Fled          bool
}

func NewPoliceEncounter() *Encounter {
	return &Encounter{
		Type:    EncPolice,
		Actions: []Action{ActionComply, ActionBribe, ActionFlee},
		Message: "Police hail your ship for inspection.",
	}
}

func NewPirateEncounter() *Encounter {
	return &Encounter{
		Type:    EncPirate,
		Actions: []Action{ActionFight, ActionFlee, ActionSurrender},
		Message: "Pirates block your path!",
	}
}

func NewTraderEncounter() *Encounter {
	return &Encounter{
		Type:    EncTrader,
		Actions: []Action{ActionTrade, ActionDecline},
		Message: "A trader offers to deal.",
	}
}
