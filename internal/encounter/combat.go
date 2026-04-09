package encounter

import (
	"fmt"
	"math/rand"

	"github.com/the4ofus/spacetrader-tui/internal/formula"
	"github.com/the4ofus/spacetrader-tui/internal/game"
	"github.com/the4ofus/spacetrader-tui/internal/gamedata"
)

type CombatRound struct {
	AttackerName string
	WeaponName   string
	Hit          bool
	RawDamage    int
	ShieldAbsorb int
	HullDamage   int
	ShieldStatus string
	HullStatus   string
}

type CombatResult struct {
	PlayerWon  bool
	PlayerFled bool
	Rounds     []CombatRound
	PlayerHull int
	EnemyHull  int
	Bounty     int
	Loot       map[int]int
}

type EnemyShip struct {
	Name         string
	Hull         int
	MaxHull      int
	Shields      []int
	WeaponPower  int
	PilotSkill   int
	FighterSkill int
	EngineerSkill int
	ShipPrice    int
	ShipSize     int
}

func NewPirateShip(gs *game.GameState) EnemyShip {
	diff := int(gs.Difficulty)
	base := 10 + diff*3
	hull := 50 + gs.Rand.Intn(50) + diff*20 + gs.Day/10
	weaponPower := base + gs.Rand.Intn(20)
	pilotSkill := 3 + gs.Rand.Intn(3) + diff
	fighterSkill := 3 + gs.Rand.Intn(3) + diff
	engineerSkill := 2 + gs.Rand.Intn(3) + diff

	var shields []int
	if diff >= 2 && gs.Rand.Intn(3) > 0 {
		shields = append(shields, 50+gs.Rand.Intn(50)+diff*25)
	}
	if diff >= 4 && gs.Rand.Intn(2) > 0 {
		shields = append(shields, 50+gs.Rand.Intn(50))
	}

	shipPrice := 10000 + gs.Rand.Intn(40000) + diff*10000
	shipSize := 1 + gs.Rand.Intn(3)

	return EnemyShip{
		Name: "Pirate", Hull: hull, MaxHull: hull,
		Shields: shields, WeaponPower: weaponPower,
		PilotSkill: pilotSkill, FighterSkill: fighterSkill,
		EngineerSkill: engineerSkill, ShipPrice: shipPrice, ShipSize: shipSize,
	}
}

func NewPoliceShip(gs *game.GameState) EnemyShip {
	diff := int(gs.Difficulty)
	hull := 100 + diff*25
	weaponPower := 20 + diff*5
	pilotSkill := 5 + diff
	fighterSkill := 5 + diff
	engineerSkill := 4 + diff

	shields := []int{100 + diff*25}
	shipPrice := 30000 + diff*10000

	return EnemyShip{
		Name: "Police", Hull: hull, MaxHull: hull,
		Shields: shields, WeaponPower: weaponPower,
		PilotSkill: pilotSkill, FighterSkill: fighterSkill,
		EngineerSkill: engineerSkill, ShipPrice: shipPrice, ShipSize: 2,
	}
}

func NewSpaceMonster(gs *game.GameState) EnemyShip {
	diff := int(gs.Difficulty)
	hull := gs.Quests.MonsterHull
	if hull <= 0 {
		hull = game.MonsterMaxHull
	}
	return EnemyShip{
		Name: "Space Monster", Hull: hull, MaxHull: game.MonsterMaxHull,
		Shields: nil, WeaponPower: 35,
		PilotSkill: 8 + diff, FighterSkill: 8 + diff,
		EngineerSkill: 1 + diff, ShipPrice: 50000, ShipSize: 3,
	}
}

func RunCombat(gs *game.GameState, enemy EnemyShip, maxRounds int) CombatResult {
	result := CombatResult{}
	rng := gs.Rand

	playerFighter := game.EffectivePlayerSkill(gs, formula.SkillFighter)
	playerPilot := game.EffectivePlayerSkill(gs, formula.SkillPilot)
	playerEngineer := game.EffectivePlayerSkill(gs, formula.SkillEngineer)
	playerShipDef := gs.Data.Ships[gs.Player.Ship.TypeID]

	playerWeapons := collectWeapons(gs)
	playerShields := collectShields(gs)
	enemyShields := make([]int, len(enemy.Shields))
	copy(enemyShields, enemy.Shields)

	for round := 0; round < maxRounds; round++ {
		playerRound := attackRound(rng, "You", playerWeapons, playerFighter, playerEngineer,
			int(playerShipDef.Size), enemy.PilotSkill, false,
			&enemyShields, &enemy.Hull, enemy.MaxHull, gs.Difficulty)
		result.Rounds = append(result.Rounds, playerRound)

		if enemy.Hull <= 0 {
			result.PlayerWon = true
			result.EnemyHull = 0
			result.PlayerHull = gs.Player.Ship.Hull
			result.Bounty = CalculateBounty(enemy.ShipPrice)

			if canScoop(rng, gs.Difficulty) {
				result.Loot = generateLoot(rng, gs)
			}
			return result
		}

		enemyRound := attackRound(rng, enemy.Name, []WeaponInfo{{Name: "Laser", Power: enemy.WeaponPower}},
			enemy.FighterSkill, enemy.EngineerSkill,
			enemy.ShipSize, playerPilot, false,
			&playerShields, &gs.Player.Ship.Hull, playerShipDef.Hull, gs.Difficulty)
		result.Rounds = append(result.Rounds, enemyRound)

		if gs.Player.Ship.Hull <= 0 {
			gs.Player.Ship.Hull = 0
			result.PlayerWon = false
			result.EnemyHull = enemy.Hull
			result.PlayerHull = 0
			return result
		}
	}

	result.PlayerWon = false
	result.EnemyHull = enemy.Hull
	result.PlayerHull = gs.Player.Ship.Hull
	return result
}

func FleeAttempt(rng *rand.Rand, playerPilot int, enemyPilot int, diff gamedata.Difficulty) bool {
	if diff == gamedata.DiffBeginner {
		return true
	}
	playerRoll := (rng.Intn(7) + playerPilot/3) * 2
	enemyRoll := rng.Intn(enemyPilot+1) * (2 + int(diff))
	return playerRoll >= enemyRoll
}

func FleeDamage(rng *rand.Rand, enemy EnemyShip, gs *game.GameState) CombatRound {
	playerPilot := game.EffectivePlayerSkill(gs, formula.SkillPilot)
	playerShields := collectShields(gs)
	playerShipDef := gs.Data.Ships[gs.Player.Ship.TypeID]

	round := attackRound(rng, enemy.Name, []WeaponInfo{{Name: "Laser", Power: enemy.WeaponPower}},
		enemy.FighterSkill, enemy.EngineerSkill,
		enemy.ShipSize, playerPilot, true,
		&playerShields, &gs.Player.Ship.Hull, playerShipDef.Hull, gs.Difficulty)
	return round
}

type WeaponInfo struct {
	Name  string
	Power int
}

func collectWeapons(gs *game.GameState) []WeaponInfo {
	var weapons []WeaponInfo
	for _, wID := range gs.Player.Ship.Weapons {
		eq := gs.Data.Equipment[wID]
		weapons = append(weapons, WeaponInfo{Name: eq.Name, Power: eq.Power})
	}
	return weapons
}

func collectShields(gs *game.GameState) []int {
	var shields []int
	for _, sID := range gs.Player.Ship.Shields {
		shields = append(shields, gs.Data.Equipment[sID].Protection)
	}
	return shields
}

func attackRound(rng *rand.Rand, attackerName string, weapons []WeaponInfo,
	fighterSkill int, engineerSkill int,
	attackerSize int, defenderPilot int, defenderFleeing bool,
	defenderShields *[]int, defenderHull *int, defenderMaxHull int,
	diff gamedata.Difficulty) CombatRound {

	if len(weapons) == 0 {
		return CombatRound{
			AttackerName: attackerName, WeaponName: "none",
			Hit: false, ShieldStatus: shieldStatusStr(*defenderShields),
			HullStatus: fmt.Sprintf("Hull: %d", *defenderHull),
		}
	}

	weapon := weapons[rng.Intn(len(weapons))]

	hitChance := fighterSkill + attackerSize
	dodgeChance := 5 + defenderPilot/2
	if defenderFleeing {
		dodgeChance /= 2
	}

	hit := rng.Intn(hitChance+dodgeChance) < hitChance

	round := CombatRound{
		AttackerName: attackerName,
		WeaponName:   weapon.Name,
		Hit:          hit,
	}

	if !hit {
		round.ShieldStatus = shieldStatusStr(*defenderShields)
		round.HullStatus = fmt.Sprintf("Hull: %d", *defenderHull)
		return round
	}

	totalWeaponPower := weapon.Power
	rawDamage := rng.Intn(totalWeaponPower*(100+2*engineerSkill)/100 + 1)
	round.RawDamage = rawDamage

	remaining := rawDamage
	absorbed := 0
	for i := range *defenderShields {
		if remaining <= 0 {
			break
		}
		absorb := min(remaining, (*defenderShields)[i])
		(*defenderShields)[i] -= absorb
		absorbed += absorb
		remaining -= absorb
	}
	round.ShieldAbsorb = absorbed

	hullDamage := remaining
	hullCap := hullDamageCap(defenderMaxHull, diff)
	if hullCap > 0 && hullDamage > hullCap {
		hullDamage = hullCap
	}

	if hullDamage > 0 {
		engReduce := rng.Intn(max(1, defenderPilot))
		hullDamage -= engReduce
		if hullDamage < 0 {
			hullDamage = 0
		}
	}

	round.HullDamage = hullDamage
	*defenderHull -= hullDamage
	if *defenderHull < 0 {
		*defenderHull = 0
	}

	round.ShieldStatus = shieldStatusStr(*defenderShields)
	round.HullStatus = fmt.Sprintf("Hull: %d", *defenderHull)

	return round
}

func hullDamageCap(maxHull int, diff gamedata.Difficulty) int {
	switch diff {
	case gamedata.DiffBeginner:
		return maxHull / 4
	case gamedata.DiffEasy:
		return maxHull / 3
	case gamedata.DiffNormal:
		return maxHull / 2
	default:
		return 0
	}
}

func shieldStatusStr(shields []int) string {
	total := 0
	for _, s := range shields {
		total += s
	}
	if total > 0 {
		return fmt.Sprintf("Shields: %d", total)
	}
	return "Shields: DOWN"
}

func CalculateBounty(shipPrice int) int {
	bounty := shipPrice / BountyDivisor
	bounty = (bounty / BountyRounding) * BountyRounding
	if bounty < BountyMin {
		bounty = BountyMin
	}
	if bounty > BountyMax {
		bounty = BountyMax
	}
	return bounty
}

func canScoop(rng *rand.Rand, diff gamedata.Difficulty) bool {
	switch diff {
	case gamedata.DiffBeginner, gamedata.DiffEasy:
		return true
	case gamedata.DiffNormal:
		return rng.Intn(2) == 0
	case gamedata.DiffHard:
		return rng.Intn(3) == 0
	default:
		return rng.Intn(4) == 0
	}
}

func generateLoot(rng *rand.Rand, gs *game.GameState) map[int]int {
	loot := map[int]int{}
	dp := &game.GameDataProvider{Data: gs.Data}
	free := gs.Player.FreeCargo(dp)
	if free <= 0 {
		return nil
	}
	goodIdx := rng.Intn(game.NumGoods)
	qty := 1 + rng.Intn(3)
	if qty > free {
		qty = free
	}
	loot[goodIdx] = qty
	gs.Player.Cargo[goodIdx] += qty
	return loot
}

func FormatCombatLog(rounds []CombatRound) string {
	log := ""
	for _, r := range rounds {
		if !r.Hit {
			log += fmt.Sprintf("%s fires %s: miss!\n", r.AttackerName, r.WeaponName)
			continue
		}
		line := fmt.Sprintf("%s fires %s: hit! %d damage", r.AttackerName, r.WeaponName, r.RawDamage)
		if r.ShieldAbsorb > 0 {
			line += fmt.Sprintf(" (shields absorb %d", r.ShieldAbsorb)
			if r.HullDamage > 0 {
				line += fmt.Sprintf(", hull takes %d", r.HullDamage)
			}
			line += ")"
		} else if r.HullDamage > 0 {
			line += fmt.Sprintf(" (hull takes %d)", r.HullDamage)
		}
		log += line + "\n"
	}
	return log
}
