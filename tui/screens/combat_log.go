package screens

import (
	"fmt"
	"strings"
	"time"

	"github.com/the4ofus/spacetrader-tui/internal/encounter"
	"github.com/the4ofus/spacetrader-tui/internal/formula"
	"github.com/the4ofus/spacetrader-tui/internal/game"
)

type CombatLogAnimator struct {
	lines     []game.CombatLogLine
	current   int
	tw        *Typewriter
	completed []string
	done      bool
}

func NewCombatLogAnimator(lines []game.CombatLogLine) *CombatLogAnimator {
	a := &CombatLogAnimator{lines: lines}
	if len(lines) == 0 {
		a.done = true
		return a
	}
	a.tw = NewTypewriter(combatLinePlain(lines[0]), AnimTypewriterEncounter)
	return a
}

func (a *CombatLogAnimator) Start(now time.Time) {
	if a.tw != nil {
		a.tw.Start(now)
	}
}

func (a *CombatLogAnimator) Update(now time.Time) {
	if a.done || a.tw == nil {
		return
	}
	a.tw.Start(now)
	a.tw.Update(now)
	if a.tw.Done() {
		a.completed = append(a.completed, combatLineStyled(a.lines[a.current]))
		a.current++
		if a.current >= len(a.lines) {
			a.done = true
			a.tw = nil
		} else {
			a.tw = NewTypewriter(combatLinePlain(a.lines[a.current]), AnimTypewriterEncounter)
		}
	}
}

func (a *CombatLogAnimator) Skip() {
	for i := a.current; i < len(a.lines); i++ {
		a.completed = append(a.completed, combatLineStyled(a.lines[i]))
	}
	a.current = len(a.lines)
	a.done = true
	a.tw = nil
}

func (a *CombatLogAnimator) Done() bool {
	return a.done
}

func (a *CombatLogAnimator) View() string {
	var b strings.Builder
	for _, line := range a.completed {
		b.WriteString("  " + line + "\n")
	}
	if a.tw != nil && !a.done {
		b.WriteString("  " + a.tw.View() + "\n")
	}
	return b.String()
}

func (a *CombatLogAnimator) StaticView() string {
	var b strings.Builder
	for _, line := range a.completed {
		b.WriteString("  " + line + "\n")
	}
	return b.String()
}

func combatLinePlain(l game.CombatLogLine) string {
	if l.InfoText != "" {
		return l.InfoText
	}
	verb := "fires"
	if l.IsPlayer {
		verb = "fire"
	}
	if l.Weapon == "" {
		if !l.Hit {
			return fmt.Sprintf("%s attack -- miss!", l.Attacker)
		}
		return fmt.Sprintf("%s hit for %d damage.", l.Attacker, l.Damage)
	}
	if !l.Hit {
		return fmt.Sprintf("%s %s %s: miss!", l.Attacker, verb, l.Weapon)
	}
	text := fmt.Sprintf("%s %s %s: hit! %d damage", l.Attacker, verb, l.Weapon, l.Damage)
	if l.ShieldAbsorb > 0 {
		text += fmt.Sprintf(" (shields absorb %d", l.ShieldAbsorb)
		if l.HullDamage > 0 {
			text += fmt.Sprintf(", hull takes %d", l.HullDamage)
		}
		text += ")"
	} else if l.HullDamage > 0 {
		text += fmt.Sprintf(" (hull takes %d)", l.HullDamage)
	}
	return text
}

func combatLineStyled(l game.CombatLogLine) string {
	if l.InfoText != "" {
		return DimStyle.Render(l.InfoText)
	}

	verb := "fires"
	if l.IsPlayer {
		verb = "fire"
	}

	if !l.Hit {
		return DimStyle.Render(combatLinePlain(l))
	}

	if l.IsPlayer && l.Hit {
		if l.Weapon == "" {
			return NormalStyle.Render("You") + " hit for " +
				SuccessStyle.Render(fmt.Sprintf("%d damage", l.Damage)) + "."
		}
		text := NormalStyle.Render("You") + " fire " + NormalStyle.Render(l.Weapon) + ": " +
			SuccessStyle.Render("hit!") + " " + SuccessStyle.Render(fmt.Sprintf("%d damage", l.Damage))
		text += combatLineShieldInfo(l)
		return text
	}

	if l.Weapon == "" {
		return DangerStyle.Render(l.Attacker) + " hits for " +
			DangerStyle.Render(fmt.Sprintf("%d damage", l.Damage)) + "."
	}
	text := DangerStyle.Render(l.Attacker) + " " + verb + " " + NormalStyle.Render(l.Weapon) + ": " +
		DangerStyle.Render("hit!") + " " + DangerStyle.Render(fmt.Sprintf("%d damage", l.Damage))
	text += combatLineShieldInfo(l)
	return text
}

func combatLineShieldInfo(l game.CombatLogLine) string {
	if l.ShieldAbsorb > 0 {
		info := " (shields absorb " + CyanStyle.Render(fmt.Sprintf("%d", l.ShieldAbsorb))
		if l.HullDamage > 0 {
			info += ", hull takes " + DangerStyle.Render(fmt.Sprintf("%d", l.HullDamage))
		}
		info += ")"
		return info
	}
	if l.HullDamage > 0 {
		return " (hull takes " + DangerStyle.Render(fmt.Sprintf("%d", l.HullDamage)) + ")"
	}
	return ""
}

func BuildCombatStatsLines(gs *game.GameState, enemy *encounter.EnemyShip) []game.CombatLogLine {
	if !VerboseCombat || enemy == nil {
		return nil
	}

	playerFighter := game.EffectivePlayerSkill(gs, formula.SkillFighter)
	playerPilot := game.EffectivePlayerSkill(gs, formula.SkillPilot)
	playerEngineer := game.EffectivePlayerSkill(gs, formula.SkillEngineer)
	shipDef := gs.PlayerShipDef()

	var weaponNames []string
	totalWeaponPower := 0
	for _, wID := range gs.Player.Ship.Weapons {
		eq := gs.Data.Equipment[wID]
		weaponNames = append(weaponNames, fmt.Sprintf("%s(%d)", eq.Name, eq.Power))
		totalWeaponPower += eq.Power
	}
	var shieldNames []string
	totalShieldHP := 0
	for _, sID := range gs.Player.Ship.Shields {
		eq := gs.Data.Equipment[sID]
		shieldNames = append(shieldNames, fmt.Sprintf("%s(%d)", eq.Name, eq.Protection))
		totalShieldHP += eq.Protection
	}

	enemyShieldTotal := 0
	for _, s := range enemy.Shields {
		enemyShieldTotal += s
	}

	var lines []game.CombatLogLine
	info := func(text string) {
		lines = append(lines, game.CombatLogLine{InfoText: text})
	}

	info(fmt.Sprintf("--- YOU: %s ---", shipDef.Name))
	info(fmt.Sprintf("Hull: %d/%d  |  Shields: %d", gs.Player.Ship.Hull, shipDef.Hull, totalShieldHP))
	if len(weaponNames) > 0 {
		info(fmt.Sprintf("Weapons: %s (total power: %d)", strings.Join(weaponNames, ", "), totalWeaponPower))
	} else {
		info("Weapons: none")
	}
	if len(shieldNames) > 0 {
		info(fmt.Sprintf("Shields: %s", strings.Join(shieldNames, ", ")))
	}
	info(fmt.Sprintf("Skills: pilot %d, fighter %d, engineer %d", playerPilot, playerFighter, playerEngineer))

	info(fmt.Sprintf("--- %s ---", strings.ToUpper(enemy.Name)))
	info(fmt.Sprintf("Hull: %d  |  Shields: %d", enemy.Hull, enemyShieldTotal))
	info(fmt.Sprintf("Weapon power: %d", enemy.WeaponPower))
	info(fmt.Sprintf("Skills: pilot %d, fighter %d, engineer %d", enemy.PilotSkill, enemy.FighterSkill, enemy.EngineerSkill))
	info("---")

	return lines
}
