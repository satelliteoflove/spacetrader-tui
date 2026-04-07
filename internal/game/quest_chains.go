package game

import (
	"fmt"

	"github.com/the4ofus/spacetrader-tui/internal/formula"
	"github.com/the4ofus/spacetrader-tui/internal/gamedata"
)

func findSystem(gs *GameState, name string) int {
	for i, sys := range gs.Data.Systems {
		if sys.Name == name {
			return i
		}
	}
	return -1
}

func findEquipByName(gs *GameState, name string) int {
	for i, eq := range gs.Data.Equipment {
		if eq.Name == name {
			return i
		}
	}
	return -1
}

func giveQuestEquipment(gs *GameState, equipName string) string {
	eqID := findEquipByName(gs, equipName)
	if eqID < 0 {
		return ""
	}
	eq := gs.Data.Equipment[eqID]
	shipDef := gs.Data.Ships[gs.Player.Ship.TypeID]

	switch eq.Category {
	case gamedata.EquipWeapon:
		if len(gs.Player.Ship.Weapons) < shipDef.WeaponSlots {
			gs.Player.Ship.Weapons = append(gs.Player.Ship.Weapons, eqID)
			return fmt.Sprintf("Received %s!", equipName)
		}
		return fmt.Sprintf("Received %s! (No weapon slot -- sell an existing weapon to install it.)", equipName)
	case gamedata.EquipShield:
		if len(gs.Player.Ship.Shields) < shipDef.ShieldSlots {
			gs.Player.Ship.Shields = append(gs.Player.Ship.Shields, eqID)
			return fmt.Sprintf("Received %s!", equipName)
		}
		return fmt.Sprintf("Received %s! (No shield slot -- sell an existing shield to install it.)", equipName)
	case gamedata.EquipGadget:
		if len(gs.Player.Ship.Gadgets) < shipDef.GadgetSlots {
			gs.Player.Ship.Gadgets = append(gs.Player.Ship.Gadgets, eqID)
			return fmt.Sprintf("Received %s!", equipName)
		}
		return fmt.Sprintf("Received %s! (No gadget slot -- sell an existing gadget to install it.)", equipName)
	}
	return ""
}

var dragonflyPath = []string{"Arouan", "Halley", "Regulus", "Linnet"}

func checkDragonfly(gs *GameState) []QuestEvent {
	state := gs.Quests.States[QuestDragonfly]
	progress := gs.Quests.Progress[QuestDragonfly]

	if state == QuestUnavailable && gs.Day > 20 && gs.Rand.Intn(100) < 10 {
		startSys := findSystem(gs, dragonflyPath[0])
		if gs.CurrentSystemID == startSys || startSys < 0 {
			return nil
		}
		cur := gs.Data.Systems[gs.CurrentSystemID]
		dest := gs.Data.Systems[startSys]
		dist := formula.Distance(cur.X, cur.Y, dest.X, dest.Y)
		gs.Quests.States[QuestDragonfly] = QuestAvailable
		return []QuestEvent{{
			Title:   "Dragonfly",
			Message: fmt.Sprintf("Reports indicate a stolen experimental ship, the Dragonfly, was last seen near %s.\n\n  First stop: %s (%.1f parsecs)\n  Route: %s -> %s -> %s -> %s\n  Reward: Lightning Shield\n  Deadline: None",
				dragonflyPath[0], dragonflyPath[0], dist,
				dragonflyPath[0], dragonflyPath[1], dragonflyPath[2], dragonflyPath[3]),
		}}
	}

	if state == QuestAvailable || state == QuestActive {
		targetName := dragonflyPath[progress]
		targetSys := findSystem(gs, targetName)
		if gs.CurrentSystemID == targetSys {
			gs.Quests.States[QuestDragonfly] = QuestActive
			progress++
			gs.Quests.Progress[QuestDragonfly] = progress
			if progress >= len(dragonflyPath) {
				gs.Quests.States[QuestDragonfly] = QuestComplete
				reward := giveQuestEquipment(gs, "Lightning Shield")
				return []QuestEvent{{
					Title:   "Dragonfly Destroyed!",
					Message: fmt.Sprintf("You destroyed the Dragonfly! %s", reward),
				}}
			}
			next := dragonflyPath[progress]
			return []QuestEvent{{
				Title:   "Dragonfly Spotted",
				Message: fmt.Sprintf("The Dragonfly was here but fled to %s!", next),
			}}
		}
	}
	return nil
}

func checkSpaceMonster(gs *GameState) []QuestEvent {
	state := gs.Quests.States[QuestSpaceMonster]

	if state == QuestUnavailable && gs.Day > 15 && gs.Rand.Intn(100) < 8 {
		acamar := findSystem(gs, "Acamar")
		distStr := ""
		if acamar >= 0 {
			cur := gs.Data.Systems[gs.CurrentSystemID]
			dest := gs.Data.Systems[acamar]
			distStr = fmt.Sprintf("\n  Location: Acamar (%.1f parsecs)", formula.Distance(cur.X, cur.Y, dest.X, dest.Y))
		}
		gs.Quests.States[QuestSpaceMonster] = QuestAvailable
		return []QuestEvent{{
			Title:   "Space Monster",
			Message: fmt.Sprintf("A terrifying Space Monster is attacking ships near Acamar! Bounty offered for its destruction.\n%s\n  Reward: 10,000 credits + reputation\n  Risk: Combat -- strength depends on fighter skill and weapons\n  Deadline: None", distStr),
		}}
	}

	acamar := findSystem(gs, "Acamar")
	if state == QuestAvailable && gs.CurrentSystemID == acamar {
		fighterSkill := EffectivePlayerSkill(gs, formula.SkillFighter)
		weaponPower := 0
		for _, w := range gs.Player.Ship.Weapons {
			weaponPower += gs.Data.Equipment[w].Power
		}
		power := fighterSkill*2 + weaponPower
		monsterPower := 30 + gs.Rand.Intn(20)

		return []QuestEvent{{
			Title:   "Space Monster!",
			Message: fmt.Sprintf("The Space Monster attacks! Your power: %d vs Monster: %d", power, monsterPower),
			Actions: []string{"Fight!", "Flee"},
		}}
	}
	return nil
}

func checkScarab(gs *GameState) []QuestEvent {
	state := gs.Quests.States[QuestScarab]

	if state == QuestUnavailable && gs.Day > 30 && gs.Rand.Intn(100) < 5 {
		gs.Quests.States[QuestScarab] = QuestAvailable
		return []QuestEvent{{
			Title:   "Scarab Sighting",
			Message: "The legendary Scarab ship has been spotted hiding near a wormhole exit. It is said to have an impenetrable hull.\n\n  Location: Near wormhole exits (chance encounter)\n  Reward: Salvaged hull plating (+20 max hull)\n  Deadline: None",
		}}
	}

	if state == QuestAvailable {
		for _, wh := range gs.Wormholes {
			if gs.CurrentSystemID == wh.SystemA || gs.CurrentSystemID == wh.SystemB {
				if gs.Rand.Intn(100) < 30 {
					return []QuestEvent{{
						Title:   "Scarab Found!",
						Message: "The Scarab is here! It's vulnerable while docked at the wormhole.",
						Actions: []string{"Attack the Scarab", "Leave it alone"},
					}}
				}
			}
		}
	}
	return nil
}

func checkAlienArtifact(gs *GameState) []QuestEvent {
	state := gs.Quests.States[QuestAlienArtifact]

	if state == QuestUnavailable && gs.Day > 25 && gs.Rand.Intn(100) < 8 {
		gs.Quests.States[QuestAlienArtifact] = QuestAvailable
		return []QuestEvent{{
			Title:   "Alien Artifact",
			Message: "You've discovered a strange alien artifact! A professor at a Hi-tech system would pay handsomely for it.\n\n  Destination: Any Hi-tech system\n  Reward: 20,000 credits + reputation\n  Risk: Mantis ships may pursue you during travel\n  Deadline: None",
			Actions: []string{"Take the artifact", "Leave it"},
		}}
	}

	if state == QuestActive {
		sys := gs.Data.Systems[gs.CurrentSystemID]
		if sys.TechLevel == gamedata.TechHiTech {
			gs.Quests.States[QuestAlienArtifact] = QuestComplete
			gs.Player.Credits += 20000
			gs.Player.Reputation += 3
			return []QuestEvent{{
				Title:   "Artifact Delivered!",
				Message: "Professor Berger is thrilled! You receive 20,000 credits and your reputation soars.",
			}}
		}
	}
	return nil
}

func checkJarek(gs *GameState) []QuestEvent {
	state := gs.Quests.States[QuestJarek]

	if state == QuestUnavailable && gs.Day > 12 && gs.Rand.Intn(100) < 10 {
		dest := findSystem(gs, "Aldebaran")
		distStr := ""
		if dest >= 0 {
			cur := gs.Data.Systems[gs.CurrentSystemID]
			d := gs.Data.Systems[dest]
			distStr = fmt.Sprintf(" (%.1f parsecs)", formula.Distance(cur.X, cur.Y, d.X, d.Y))
		}
		gs.Quests.States[QuestJarek] = QuestAvailable
		return []QuestEvent{{
			Title:   "Ambassador Jarek",
			Message: fmt.Sprintf("Ambassador Jarek needs transport to Aldebaran for urgent diplomatic negotiations.\n\n  Destination: Aldebaran%s\n  Reward: 5,000 credits + reputation\n  Deadline: 10 stops -- he leaves if not delivered in time\n  Failure: No penalty, ambassador departs", distStr),
			Actions: []string{"Accept passenger", "Decline"},
		}}
	}

	if state == QuestActive {
		aldebaran := findSystem(gs, "Aldebaran")
		if aldebaran >= 0 && gs.CurrentSystemID == aldebaran {
			gs.Quests.States[QuestJarek] = QuestComplete
			gs.Player.Credits += 5000
			gs.Player.Reputation += 2
			return []QuestEvent{{
				Title:   "Ambassador Delivered!",
				Message: "Ambassador Jarek thanks you. 5,000 credits and increased reputation.",
			}}
		}

		gs.Quests.Progress[QuestJarek]++
		if gs.Quests.Progress[QuestJarek] > 10 {
			gs.Quests.States[QuestJarek] = QuestUnavailable
			gs.Quests.Progress[QuestJarek] = 0
			return []QuestEvent{{
				Title:   "Ambassador Impatient",
				Message: "Ambassador Jarek has lost patience and left your ship at the next port.",
			}}
		}
	}
	return nil
}

func checkGemulon(gs *GameState) []QuestEvent {
	state := gs.Quests.States[QuestGemulon]

	if state == QuestUnavailable && gs.Day > 35 && gs.Rand.Intn(100) < 6 {
		gemulon := findSystem(gs, "Gemulon")
		distStr := ""
		if gemulon >= 0 {
			cur := gs.Data.Systems[gs.CurrentSystemID]
			dest := gs.Data.Systems[gemulon]
			distStr = fmt.Sprintf(" (%.1f parsecs)", formula.Distance(cur.X, cur.Y, dest.X, dest.Y))
		}
		gs.Quests.States[QuestGemulon] = QuestAvailable
		gs.Quests.Progress[QuestGemulon] = gs.Day
		return []QuestEvent{{
			Title:   "Gemulon Invasion!",
			Message: fmt.Sprintf("Aliens are planning to invade Gemulon! You must warn them within 7 days!\n\n  Destination: Gemulon%s\n  Reward: Fuel Compactor\n  Deadline: 7 days -- system is invaded if you're late\n  Failure: Gemulon falls, quest lost", distStr),
		}}
	}

	if state == QuestAvailable {
		startDay := gs.Quests.Progress[QuestGemulon]
		gemulon := findSystem(gs, "Gemulon")
		if gs.Day-startDay > 7 {
			gs.Quests.States[QuestGemulon] = QuestUnavailable
			gs.Quests.Progress[QuestGemulon] = 0
			return []QuestEvent{{
				Title:   "Gemulon Invaded",
				Message: "You arrived too late. Gemulon has been invaded.",
			}}
		}
		if gemulon >= 0 && gs.CurrentSystemID == gemulon {
			gs.Quests.States[QuestGemulon] = QuestComplete
			reward := giveQuestEquipment(gs, "Fuel Compactor")
			return []QuestEvent{{
				Title:   "Gemulon Saved!",
				Message: fmt.Sprintf("You warned Gemulon in time! %s", reward),
			}}
		}
	}
	return nil
}

func checkFehler(gs *GameState) []QuestEvent {
	state := gs.Quests.States[QuestFehler]

	if state == QuestUnavailable && gs.Day > 40 && gs.Rand.Intn(100) < 5 {
		deneb := findSystem(gs, "Deneb")
		distStr := ""
		if deneb >= 0 {
			cur := gs.Data.Systems[gs.CurrentSystemID]
			dest := gs.Data.Systems[deneb]
			distStr = fmt.Sprintf(" (%.1f parsecs)", formula.Distance(cur.X, cur.Y, dest.X, dest.Y))
		}
		gs.Quests.States[QuestFehler] = QuestAvailable
		gs.Quests.Progress[QuestFehler] = gs.Day
		return []QuestEvent{{
			Title:   "Dr. Fehler's Experiment",
			Message: fmt.Sprintf("Dr. Fehler's experiment at Deneb is about to rip a hole in spacetime! Someone must stop it within 5 days!\n\n  Destination: Deneb%s\n  Reward: 10,000 credits + reputation\n  Deadline: 5 days -- spacetime distorted if you're late\n  Failure: Quest lost, no penalty", distStr),
		}}
	}

	if state == QuestAvailable {
		startDay := gs.Quests.Progress[QuestFehler]
		deneb := findSystem(gs, "Deneb")

		if gs.Day-startDay > 5 {
			gs.Quests.States[QuestFehler] = QuestUnavailable
			gs.Quests.Progress[QuestFehler] = 0
			return []QuestEvent{{
				Title:   "Experiment Failed",
				Message: "The experiment went wrong. Spacetime is distorted near Deneb.",
			}}
		}
		if deneb >= 0 && gs.CurrentSystemID == deneb {
			gs.Quests.States[QuestFehler] = QuestComplete
			gs.Player.Credits += 10000
			gs.Player.Reputation += 3
			return []QuestEvent{{
				Title:   "Experiment Stopped!",
				Message: "You stopped Dr. Fehler's experiment just in time! 10,000 credits reward.",
			}}
		}
	}
	return nil
}

func checkWild(gs *GameState) []QuestEvent {
	state := gs.Quests.States[QuestWild]

	if state == QuestUnavailable && gs.Day > 18 && gs.Rand.Intn(100) < 7 {
		sys := gs.Data.Systems[gs.CurrentSystemID]
		if sys.PoliticalSystem == gamedata.PolAnarchy || sys.PoliticalSystem == gamedata.PolFeudal {
			adahn := findSystem(gs, "Adahn")
			distStr := ""
			if adahn >= 0 {
				cur := gs.Data.Systems[gs.CurrentSystemID]
				dest := gs.Data.Systems[adahn]
				distStr = fmt.Sprintf(" (%.1f parsecs)", formula.Distance(cur.X, cur.Y, dest.X, dest.Y))
			}
			gs.Quests.States[QuestWild] = QuestAvailable
			return []QuestEvent{{
				Title:   "Jonathan Wild",
				Message: fmt.Sprintf("The notorious criminal Jonathan Wild wants passage to Adahn.\n\n  Destination: Adahn%s\n  Reward: 15,000 credits\n  Risk: Police record worsens (-5)\n  Deadline: None", distStr),
				Actions: []string{"Take him aboard", "Refuse"},
			}}
		}
	}

	if state == QuestActive {
		adahn := findSystem(gs, "Adahn")
		if adahn >= 0 && gs.CurrentSystemID == adahn {
			gs.Quests.States[QuestWild] = QuestComplete
			gs.Player.Credits += 15000
			gs.Player.PoliceRecord -= 5
			return []QuestEvent{{
				Title:   "Wild Delivered",
				Message: "Jonathan Wild disappears into the crowd. 15,000 credits, but your record suffers.",
			}}
		}
	}
	return nil
}

func checkReactor(gs *GameState) []QuestEvent {
	state := gs.Quests.States[QuestReactor]

	if state == QuestUnavailable && gs.Day > 45 && gs.Rand.Intn(100) < 5 {
		dp := &GameDataProvider{Data: gs.Data}
		if gs.Player.FreeCargo(dp) >= 5 {
			eridani := findSystem(gs, "Eridani")
			distStr := ""
			if eridani >= 0 {
				cur := gs.Data.Systems[gs.CurrentSystemID]
				dest := gs.Data.Systems[eridani]
				distStr = fmt.Sprintf(" (%.1f parsecs)", formula.Distance(cur.X, cur.Y, dest.X, dest.Y))
			}
			gs.Quests.States[QuestReactor] = QuestAvailable
			return []QuestEvent{{
				Title:   "Reactor Delivery",
				Message: fmt.Sprintf("Henry Morgan needs an unstable reactor delivered to Eridani.\n\n  Destination: Eridani%s\n  Reward: Morgan's Laser\n  Cost: 5 cargo bays while carrying\n  Risk: Reactor leaks fuel (-1 per stop)\n  Deadline: None", distStr),
				Actions: []string{"Accept the reactor", "Decline"},
			}}
		}
	}

	if state == QuestActive {
		gs.Player.Ship.Fuel -= 1
		if gs.Player.Ship.Fuel < 0 {
			gs.Player.Ship.Fuel = 0
		}

		eridani := findSystem(gs, "Eridani")
		if eridani >= 0 && gs.CurrentSystemID == eridani {
			gs.Quests.States[QuestReactor] = QuestComplete
			gs.Quests.Progress[QuestReactor] = 0
			reward := giveQuestEquipment(gs, "Morgan's Laser")
			return []QuestEvent{{
				Title:   "Reactor Delivered!",
				Message: fmt.Sprintf("Henry Morgan is pleased. %s", reward),
			}}
		}
	}
	return nil
}

func resolveQuestChainAction(gs *GameState, title string, actionIdx int) string {
	switch title {
	case "Space Monster!":
		if actionIdx == 0 {
			fighterSkill := EffectivePlayerSkill(gs, formula.SkillFighter)
			weaponPower := 0
			for _, w := range gs.Player.Ship.Weapons {
				weaponPower += gs.Data.Equipment[w].Power
			}
			power := fighterSkill*2 + weaponPower
			monsterPower := 30 + gs.Rand.Intn(20)

			if power >= monsterPower {
				gs.Quests.States[QuestSpaceMonster] = QuestComplete
				gs.Player.Credits += 10000
				gs.Player.Reputation += 5
				return "You destroyed the Space Monster! 10,000 credits bounty and fame across the galaxy!"
			}
			damage := 20 + gs.Rand.Intn(30)
			gs.Player.Ship.Hull -= damage
			if gs.Player.Ship.Hull < 1 {
				gs.Player.Ship.Hull = 1
			}
			return fmt.Sprintf("The monster is too powerful! You barely escape with %d hull damage.", damage)
		}
		return "You flee from the Space Monster."

	case "Scarab Found!":
		if actionIdx == 0 {
			gs.Quests.States[QuestScarab] = QuestComplete
			shipDef := gs.Data.Ships[gs.Player.Ship.TypeID]
			gs.Player.Ship.Hull = shipDef.Hull + 20
			return "You destroyed the Scarab! Its hull plating was salvaged -- your hull is permanently reinforced (+20 max hull)."
		}
		return "You leave the Scarab alone."

	case "Alien Artifact":
		if actionIdx == 0 {
			gs.Quests.States[QuestAlienArtifact] = QuestActive
			return "You take the alien artifact. Mantis ships may now pursue you during travel."
		}
		gs.Quests.States[QuestAlienArtifact] = QuestUnavailable
		return "You leave the artifact."

	case "Ambassador Jarek":
		if actionIdx == 0 {
			gs.Quests.States[QuestJarek] = QuestActive
			gs.Quests.Progress[QuestJarek] = 0
			return "Ambassador Jarek boards your ship. Deliver him to Aldebaran."
		}
		gs.Quests.States[QuestJarek] = QuestUnavailable
		return "Declined."

	case "Jonathan Wild":
		if actionIdx == 0 {
			gs.Quests.States[QuestWild] = QuestActive
			return "Jonathan Wild is now aboard. Deliver him to Adahn -- but watch out for police."
		}
		gs.Quests.States[QuestWild] = QuestUnavailable
		return "You refuse to smuggle a criminal."

	case "Reactor Delivery":
		if actionIdx == 0 {
			dp := &GameDataProvider{Data: gs.Data}
			if gs.Player.FreeCargo(dp) < 5 {
				return "Not enough cargo space for the reactor (need 5 free bays)."
			}
			gs.Quests.States[QuestReactor] = QuestActive
			gs.Quests.Progress[QuestReactor] = 5
			return "Reactor loaded (5 cargo bays). It will slowly leak fuel. Deliver to Eridani."
		}
		gs.Quests.States[QuestReactor] = QuestUnavailable
		return "Declined."
	}
	return ""
}
