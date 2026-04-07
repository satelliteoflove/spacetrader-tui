package gamedata

import "fmt"

type TechLevel int

const (
	TechPreAgricultural TechLevel = iota
	TechAgricultural
	TechMedieval
	TechRenaissance
	TechEarlyIndustrial
	TechIndustrial
	TechPostIndustrial
	TechHiTech
	NumTechLevels
)

var techLevelNames = [NumTechLevels]string{
	"Pre-agricultural",
	"Agricultural",
	"Medieval",
	"Renaissance",
	"Early Industrial",
	"Industrial",
	"Post-industrial",
	"Hi-tech",
}

func (t TechLevel) String() string {
	if t >= 0 && t < NumTechLevels {
		return techLevelNames[t]
	}
	return fmt.Sprintf("TechLevel(%d)", int(t))
}

func ParseTechLevel(s string) (TechLevel, error) {
	for i, name := range techLevelNames {
		if name == s {
			return TechLevel(i), nil
		}
	}
	return 0, fmt.Errorf("unknown tech level: %q", s)
}

type PoliticalSystem int

const (
	PolAnarchy PoliticalSystem = iota
	PolCapitalist
	PolCommunist
	PolConfederacy
	PolCorporate
	PolCybernetic
	PolDemocracy
	PolDictatorship
	PolFascist
	PolFeudal
	PolMilitary
	PolMonarchy
	PolPacifist
	PolSocialist
	PolSatori
	PolTechnocracy
	PolTheocracy
	NumPoliticalSystems
)

var politicalSystemNames = [NumPoliticalSystems]string{
	"Anarchy",
	"Capitalist State",
	"Communist State",
	"Confederacy",
	"Corporate State",
	"Cybernetic State",
	"Democracy",
	"Dictatorship",
	"Fascist State",
	"Feudal State",
	"Military State",
	"Monarchy",
	"Pacifist State",
	"Socialist State",
	"State of Satori",
	"Technocracy",
	"Theocracy",
}

func (p PoliticalSystem) String() string {
	if p >= 0 && p < NumPoliticalSystems {
		return politicalSystemNames[p]
	}
	return fmt.Sprintf("PoliticalSystem(%d)", int(p))
}

func ParsePoliticalSystem(s string) (PoliticalSystem, error) {
	for i, name := range politicalSystemNames {
		if name == s {
			return PoliticalSystem(i), nil
		}
	}
	return 0, fmt.Errorf("unknown political system: %q", s)
}

type Resource int

const (
	ResourceNone Resource = iota
	ResourceDesert
	ResourceMineralRich
	ResourceIndustrial
	ResourceWaterWorld
	ResourcePoor
	ResourceRichFauna
	ResourceRichSoil
	ResourceLifeless
	ResourcePoorSoil
	ResourcePoorClinic
	ResourceGoodClinic
	ResourceLackOfWorkers
	ResourceRobotWorkers
	NumResources
)

var resourceNames = [NumResources]string{
	"No Special Resources",
	"Desert",
	"Mineral Rich",
	"Industrial",
	"Water World",
	"Poor",
	"Rich Fauna",
	"Rich Soil",
	"Lifeless",
	"Poor Soil",
	"Poor Clinic",
	"Good Clinic",
	"Lack of Workers",
	"Robot Workers",
}

func (r Resource) String() string {
	if r >= 0 && r < NumResources {
		return resourceNames[r]
	}
	return fmt.Sprintf("Resource(%d)", int(r))
}

func ParseResource(s string) (Resource, error) {
	for i, name := range resourceNames {
		if name == s {
			return Resource(i), nil
		}
	}
	return 0, fmt.Errorf("unknown resource: %q", s)
}

type GoodID int

const (
	GoodWater GoodID = iota
	GoodFurs
	GoodFood
	GoodOre
	GoodGames
	GoodFirearms
	GoodMedicine
	GoodMachines
	GoodNarcotics
	GoodRobots
	NumGoods
)

var goodNames = [NumGoods]string{
	"Water",
	"Furs",
	"Food",
	"Ore",
	"Games",
	"Firearms",
	"Medicine",
	"Machines",
	"Narcotics",
	"Robots",
}

func (g GoodID) String() string {
	if g >= 0 && g < NumGoods {
		return goodNames[g]
	}
	return fmt.Sprintf("GoodID(%d)", int(g))
}

func ParseGoodID(s string) (GoodID, error) {
	for i, name := range goodNames {
		if name == s {
			return GoodID(i), nil
		}
	}
	return 0, fmt.Errorf("unknown good: %q", s)
}

type ShipType int

const (
	ShipFlea ShipType = iota
	ShipGnat
	ShipFirefly
	ShipMosquito
	ShipBumblebee
	ShipBeetle
	ShipHornet
	ShipGrasshopper
	ShipTermite
	ShipWasp
	NumShipTypes
)

var shipTypeNames = [NumShipTypes]string{
	"Flea",
	"Gnat",
	"Firefly",
	"Mosquito",
	"Bumblebee",
	"Beetle",
	"Hornet",
	"Grasshopper",
	"Termite",
	"Wasp",
}

func (s ShipType) String() string {
	if s >= 0 && s < NumShipTypes {
		return shipTypeNames[s]
	}
	return fmt.Sprintf("ShipType(%d)", int(s))
}

func ParseShipType(name string) (ShipType, error) {
	for i, n := range shipTypeNames {
		if n == name {
			return ShipType(i), nil
		}
	}
	return 0, fmt.Errorf("unknown ship type: %q", name)
}

type ShipSize int

const (
	SizeTiny ShipSize = iota
	SizeSmall
	SizeMedium
	SizeLarge
	SizeHuge
)

var shipSizeNames = map[string]ShipSize{
	"Tiny":   SizeTiny,
	"Small":  SizeSmall,
	"Medium": SizeMedium,
	"Large":  SizeLarge,
	"Huge":   SizeHuge,
}

func ParseShipSize(s string) (ShipSize, error) {
	if v, ok := shipSizeNames[s]; ok {
		return v, nil
	}
	return 0, fmt.Errorf("unknown ship size: %q", s)
}

type SystemSize int

const (
	SysTiny SystemSize = iota
	SysSmall
	SysMedium
	SysLarge
	SysHuge
	NumSystemSizes
)

var systemSizeNames = [NumSystemSizes]string{
	"Tiny",
	"Small",
	"Medium",
	"Large",
	"Huge",
}

func (s SystemSize) String() string {
	if s >= 0 && s < NumSystemSizes {
		return systemSizeNames[s]
	}
	return fmt.Sprintf("SystemSize(%d)", int(s))
}

type EquipCategory int

const (
	EquipWeapon EquipCategory = iota
	EquipShield
	EquipGadget
)

func ParseEquipCategory(s string) (EquipCategory, error) {
	switch s {
	case "weapon":
		return EquipWeapon, nil
	case "shield":
		return EquipShield, nil
	case "gadget":
		return EquipGadget, nil
	}
	return 0, fmt.Errorf("unknown equipment category: %q", s)
}

type Difficulty int

const (
	DiffBeginner Difficulty = iota
	DiffEasy
	DiffNormal
	DiffHard
	DiffImpossible
)

func (d Difficulty) String() string {
	switch d {
	case DiffBeginner:
		return "Beginner"
	case DiffEasy:
		return "Easy"
	case DiffNormal:
		return "Normal"
	case DiffHard:
		return "Hard"
	case DiffImpossible:
		return "Impossible"
	}
	return fmt.Sprintf("Difficulty(%d)", int(d))
}

type PoliceRecordTier int

const (
	RecordPsychopath PoliceRecordTier = iota
	RecordVillain
	RecordCriminal
	RecordCrook
	RecordDubious
	RecordClean
	RecordLawful
	RecordTrusted
	RecordLiked
	RecordHero
)

func PoliceRecordToTier(record int) PoliceRecordTier {
	switch {
	case record < -100:
		return RecordPsychopath
	case record < -70:
		return RecordVillain
	case record < -30:
		return RecordCriminal
	case record < -10:
		return RecordCrook
	case record < 0:
		return RecordDubious
	case record < 10:
		return RecordClean
	case record < 30:
		return RecordLawful
	case record < 70:
		return RecordTrusted
	case record < 100:
		return RecordLiked
	default:
		return RecordHero
	}
}

func (p PoliceRecordTier) String() string {
	switch p {
	case RecordPsychopath:
		return "Psychopath"
	case RecordVillain:
		return "Villain"
	case RecordCriminal:
		return "Criminal"
	case RecordCrook:
		return "Crook"
	case RecordDubious:
		return "Dubious"
	case RecordClean:
		return "Clean"
	case RecordLawful:
		return "Lawful"
	case RecordTrusted:
		return "Trusted"
	case RecordLiked:
		return "Liked"
	case RecordHero:
		return "Hero"
	}
	return "Unknown"
}

type ReputationTier int

const (
	RepHarmless ReputationTier = iota
	RepMostlyHarmless
	RepPoor
	RepAverage
	RepAboveAverage
	RepCompetent
	RepDangerous
	RepDeadly
	RepElite
)

func ReputationToTier(rep int) ReputationTier {
	switch {
	case rep < 1:
		return RepHarmless
	case rep < 3:
		return RepMostlyHarmless
	case rep < 7:
		return RepPoor
	case rep < 15:
		return RepAverage
	case rep < 25:
		return RepAboveAverage
	case rep < 50:
		return RepCompetent
	case rep < 100:
		return RepDangerous
	case rep < 200:
		return RepDeadly
	default:
		return RepElite
	}
}

func (r ReputationTier) String() string {
	switch r {
	case RepHarmless:
		return "Harmless"
	case RepMostlyHarmless:
		return "Mostly Harmless"
	case RepPoor:
		return "Poor"
	case RepAverage:
		return "Average"
	case RepAboveAverage:
		return "Above Average"
	case RepCompetent:
		return "Competent"
	case RepDangerous:
		return "Dangerous"
	case RepDeadly:
		return "Deadly"
	case RepElite:
		return "Elite"
	}
	return "Unknown"
}
