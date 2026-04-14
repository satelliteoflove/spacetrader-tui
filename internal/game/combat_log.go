package game

type CombatLogLine struct {
	Attacker     string
	Weapon       string
	Hit          bool
	Damage       int
	ShieldAbsorb int
	HullDamage   int
	IsPlayer     bool
	InfoText     string
}
