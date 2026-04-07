package gamedata

type PoliticalSystemData struct {
	Name           string
	MinTech        TechLevel
	MaxTech        TechLevel
	BribeLevel     int
	PoliceStrength int
	PirateStrength int
	TraderStrength int
	DrugLegal      bool
	FirearmLegal   bool
	WantedGood     GoodID
	HasWantedGood  bool
	NewspaperNames [3]string
}

var PoliticalSystems = [NumPoliticalSystems]PoliticalSystemData{
	{ // Anarchy
		Name: "Anarchy", MinTech: TechPreAgricultural, MaxTech: TechPreAgricultural,
		BribeLevel: 5, PoliceStrength: 1, PirateStrength: 7, TraderStrength: 0,
		DrugLegal: true, FirearmLegal: true,
		WantedGood: GoodFood, HasWantedGood: true,
		NewspaperNames: [3]string{"The Arsenal", "The Grassroot", "Kick It!"},
	},
	{ // Capitalist State
		Name: "Capitalist State", MinTech: TechMedieval, MaxTech: TechRenaissance,
		BribeLevel: 7, PoliceStrength: 2, PirateStrength: 7, TraderStrength: 4,
		DrugLegal: true, FirearmLegal: true,
		WantedGood: GoodOre, HasWantedGood: true,
		NewspaperNames: [3]string{"The Objectivist", "The Market", "The Invisible Hand"},
	},
	{ // Communist State
		Name: "Communist State", MinTech: TechPostIndustrial, MaxTech: TechPostIndustrial,
		BribeLevel: 5, PoliceStrength: 4, PirateStrength: 4, TraderStrength: 1,
		DrugLegal: true, FirearmLegal: true,
		WantedGood: 0, HasWantedGood: false,
		NewspaperNames: [3]string{"The Daily Worker", "The People's Voice", "The Proletariat"},
	},
	{ // Confederacy
		Name: "Confederacy", MinTech: TechIndustrial, MaxTech: TechEarlyIndustrial,
		BribeLevel: 6, PoliceStrength: 3, PirateStrength: 5, TraderStrength: 1,
		DrugLegal: true, FirearmLegal: true,
		WantedGood: GoodGames, HasWantedGood: true,
		NewspaperNames: [3]string{"Planet News", "The Times", "Interstate Update"},
	},
	{ // Corporate State
		Name: "Corporate State", MinTech: TechMedieval, MaxTech: TechPostIndustrial,
		BribeLevel: 7, PoliceStrength: 2, PirateStrength: 7, TraderStrength: 4,
		DrugLegal: true, FirearmLegal: true,
		WantedGood: GoodRobots, HasWantedGood: true,
		NewspaperNames: [3]string{"Memo", "News From The Board", "Status Report"},
	},
	{ // Cybernetic State
		Name: "Cybernetic State", MinTech: TechPreAgricultural, MaxTech: TechHiTech,
		BribeLevel: 7, PoliceStrength: 7, PirateStrength: 5, TraderStrength: 6,
		DrugLegal: false, FirearmLegal: false,
		WantedGood: GoodOre, HasWantedGood: true,
		NewspaperNames: [3]string{"Pulses", "Binary Stream", "The System Clock"},
	},
	{ // Democracy
		Name: "Democracy", MinTech: TechEarlyIndustrial, MaxTech: TechRenaissance,
		BribeLevel: 7, PoliceStrength: 2, PirateStrength: 5, TraderStrength: 3,
		DrugLegal: true, FirearmLegal: true,
		WantedGood: GoodGames, HasWantedGood: true,
		NewspaperNames: [3]string{"The Daily Planet", "The Majority", "Unanimity"},
	},
	{ // Dictatorship
		Name: "Dictatorship", MinTech: TechRenaissance, MaxTech: TechEarlyIndustrial,
		BribeLevel: 7, PoliceStrength: 5, PirateStrength: 3, TraderStrength: 0,
		DrugLegal: true, FirearmLegal: true,
		WantedGood: 0, HasWantedGood: false,
		NewspaperNames: [3]string{"The Command", "Leader's Voice", "The Mandate"},
	},
	{ // Fascist State
		Name: "Fascist State", MinTech: TechHiTech, MaxTech: TechHiTech,
		BribeLevel: 0, PoliceStrength: 7, PirateStrength: 1, TraderStrength: 4,
		DrugLegal: false, FirearmLegal: true,
		WantedGood: GoodMachines, HasWantedGood: true,
		NewspaperNames: [3]string{"State Tribune", "Motherland News", "Homeland Report"},
	},
	{ // Feudal State
		Name: "Feudal State", MinTech: TechAgricultural, MaxTech: TechAgricultural,
		BribeLevel: 3, PoliceStrength: 6, PirateStrength: 2, TraderStrength: 0,
		DrugLegal: true, FirearmLegal: true,
		WantedGood: GoodFirearms, HasWantedGood: true,
		NewspaperNames: [3]string{"News from the Keep", "The Town Crier", "The Herald"},
	},
	{ // Military State
		Name: "Military State", MinTech: TechHiTech, MaxTech: TechHiTech,
		BribeLevel: 0, PoliceStrength: 0, PirateStrength: 6, TraderStrength: 2,
		DrugLegal: false, FirearmLegal: true,
		WantedGood: GoodRobots, HasWantedGood: true,
		NewspaperNames: [3]string{"General Report", "Dispatch", "The Sentry"},
	},
	{ // Monarchy
		Name: "Monarchy", MinTech: TechRenaissance, MaxTech: TechEarlyIndustrial,
		BribeLevel: 5, PoliceStrength: 3, PirateStrength: 4, TraderStrength: 0,
		DrugLegal: true, FirearmLegal: true,
		WantedGood: GoodMedicine, HasWantedGood: true,
		NewspaperNames: [3]string{"Royal Times", "The Loyal Subject", "The Fanfare"},
	},
	{ // Pacifist State
		Name: "Pacifist State", MinTech: TechHiTech, MaxTech: TechMedieval,
		BribeLevel: 3, PoliceStrength: 1, PirateStrength: 5, TraderStrength: 0,
		DrugLegal: true, FirearmLegal: false,
		WantedGood: 0, HasWantedGood: false,
		NewspaperNames: [3]string{"Pax Humani", "Principle", "The Chorus"},
	},
	{ // Socialist State
		Name: "Socialist State", MinTech: TechEarlyIndustrial, MaxTech: TechMedieval,
		BribeLevel: 5, PoliceStrength: 5, PirateStrength: 3, TraderStrength: 0,
		DrugLegal: true, FirearmLegal: true,
		WantedGood: 0, HasWantedGood: false,
		NewspaperNames: [3]string{"All for One", "Brotherhood", "The People's Syndicate"},
	},
	{ // State of Satori
		Name: "State of Satori", MinTech: TechPreAgricultural, MaxTech: TechAgricultural,
		BribeLevel: 1, PoliceStrength: 1, PirateStrength: 1, TraderStrength: 0,
		DrugLegal: false, FirearmLegal: false,
		WantedGood: 0, HasWantedGood: false,
		NewspaperNames: [3]string{"The Daily Koan", "Haiku", "One Hand Clapping"},
	},
	{ // Technocracy
		Name: "Technocracy", MinTech: TechAgricultural, MaxTech: TechPostIndustrial,
		BribeLevel: 0, PoliceStrength: 3, PirateStrength: 6, TraderStrength: 4,
		DrugLegal: true, FirearmLegal: true,
		WantedGood: GoodWater, HasWantedGood: true,
		NewspaperNames: [3]string{"The Future", "Hardware Dispatch", "TechNews"},
	},
	{ // Theocracy
		Name: "Theocracy", MinTech: TechIndustrial, MaxTech: TechPostIndustrial,
		BribeLevel: 0, PoliceStrength: 1, PirateStrength: 4, TraderStrength: 0,
		DrugLegal: true, FirearmLegal: true,
		WantedGood: GoodNarcotics, HasWantedGood: true,
		NewspaperNames: [3]string{"The Spiritual Advisor", "Church Tidings", "Temple Tribune"},
	},
}
