package gamedata

type SystemDef struct {
	ID              int
	Name            string
	X, Y            int
	TechLevel       TechLevel
	PoliticalSystem PoliticalSystem
	Resource        Resource
	Size            SystemSize
	Special         string
}

type GoodDef struct {
	ID                 GoodID
	Name               string
	BasePrice          int
	Legal              bool
	MinTech            TechLevel
	MaxTech            TechLevel
	Variance           int
	PriceIncreaseEvent string
	PriceDecreaseEvent string
	ExpensiveResource  string
	CheapResource      string
}

type ShipDef struct {
	ID           int
	Name         string
	Size         ShipSize
	CargoBays    int
	WeaponSlots  int
	ShieldSlots  int
	GadgetSlots  int
	CrewQuarters int
	Range        int
	FuelCost     int
	Hull         int
	RepairCost   int
	Price        int
	MinTech      TechLevel
}

type EquipDef struct {
	ID         int
	Name       string
	Category   EquipCategory
	Power      int
	Protection int
	SkillBonus string
	CargoBays  int
	RangeBonus int
	TechLevel  TechLevel
	Price      int
	QuestReward bool
}

type GameData struct {
	Systems   []SystemDef
	Goods     []GoodDef
	Ships     []ShipDef
	Equipment []EquipDef
}
