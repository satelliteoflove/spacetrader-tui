package game

type Ship struct {
	TypeID       int   `json:"type_id"`
	Hull         int   `json:"hull"`
	Fuel         int   `json:"fuel"`
	Weapons      []int `json:"weapons"`
	Shields      []int `json:"shields"`
	Gadgets      []int `json:"gadgets"`
	HullUpgraded bool  `json:"hull_upgraded"`
}

const ScarabHullBonus = 50
