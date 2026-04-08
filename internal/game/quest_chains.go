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

func systemDistanceStr(gs *GameState, name string) (int, string) {
	idx := findSystem(gs, name)
	if idx < 0 {
		return -1, ""
	}
	cur := gs.Data.Systems[gs.CurrentSystemID]
	dest := gs.Data.Systems[idx]
	dist := formula.Distance(cur.X, cur.Y, dest.X, dest.Y)
	return idx, fmt.Sprintf(" (%.1f parsecs)", dist)
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
	state := gs.QuestState(QuestDragonfly)
	progress := gs.QuestProgress(QuestDragonfly)

	if state == QuestUnavailable && gs.Day > 20 && gs.Rand.Intn(100) < 10 {
		startSys := findSystem(gs, dragonflyPath[0])
		if gs.CurrentSystemID == startSys || startSys < 0 {
			return nil
		}
		cur := gs.Data.Systems[gs.CurrentSystemID]
		dest := gs.Data.Systems[startSys]
		dist := formula.Distance(cur.X, cur.Y, dest.X, dest.Y)
		gs.SetQuestState(QuestDragonfly, QuestAvailable)
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
			gs.SetQuestState(QuestDragonfly, QuestActive)
			progress++
			gs.SetQuestProgress(QuestDragonfly, progress)
			if progress >= len(dragonflyPath) {
				gs.SetQuestState(QuestDragonfly, QuestComplete)
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
	state := gs.QuestState(QuestSpaceMonster)

	if state == QuestUnavailable && gs.Day > 15 && gs.Rand.Intn(100) < 8 {
		_, distStr := systemDistanceStr(gs, "Acamar")
		if distStr != "" {
			distStr = "\n  Location: Acamar" + distStr
		}
		gs.SetQuestState(QuestSpaceMonster, QuestAvailable)
		return []QuestEvent{{
			Title:   "Space Monster",
			Message: fmt.Sprintf("A terrifying Space Monster is attacking ships near Acamar! Bounty offered for its destruction.\n%s\n  Reward: 10,000 credits + reputation\n  Risk: Combat -- strength depends on fighter skill and weapons\n  Deadline: None", distStr),
		}}
	}

	acamar := findSystem(gs, "Acamar")
	if state == QuestAvailable && gs.CurrentSystemID == acamar {
		monsterHull := gs.Quests.MonsterHull
		if monsterHull == 0 {
			monsterHull = MonsterMaxHull
		}

		return []QuestEvent{{
			Title:   "Space Monster!",
			Message: fmt.Sprintf("The Space Monster attacks! Monster hull: %d/%d. It regenerates 5%% hull per day -- don't delay!", monsterHull, MonsterMaxHull),
			Actions: []string{"Fight!", "Flee"},
		}}
	}
	return nil
}

func checkScarab(gs *GameState) []QuestEvent {
	state := gs.QuestState(QuestScarab)

	if state == QuestUnavailable && gs.Day > 30 && gs.Rand.Intn(100) < 5 {
		gs.SetQuestState(QuestScarab, QuestAvailable)
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
	state := gs.QuestState(QuestAlienArtifact)

	if state == QuestUnavailable && gs.Day > 25 && gs.Rand.Intn(100) < 8 {
		gs.SetQuestState(QuestAlienArtifact, QuestAvailable)
		return []QuestEvent{{
			Title:   "Alien Artifact",
			Message: "You've discovered a strange alien artifact! A professor at a Hi-tech system would pay handsomely for it.\n\n  Destination: Any Hi-tech system\n  Reward: 20,000 credits + reputation\n  Risk: Mantis ships may pursue you during travel\n  Deadline: None",
			Actions: []string{"Take the artifact", "Leave it"},
		}}
	}

	if state == QuestActive {
		sys := gs.Data.Systems[gs.CurrentSystemID]
		if sys.TechLevel == gamedata.TechHiTech {
			gs.SetQuestState(QuestAlienArtifact, QuestComplete)
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
	state := gs.QuestState(QuestJarek)

	if state == QuestUnavailable && gs.Day > 12 && gs.Rand.Intn(100) < 10 {
		_, distStr := systemDistanceStr(gs, "Aldebaran")
		gs.SetQuestState(QuestJarek, QuestAvailable)
		return []QuestEvent{{
			Title:   "Ambassador Jarek",
			Message: fmt.Sprintf("Ambassador Jarek needs transport to Aldebaran for urgent diplomatic negotiations.\n\n  Destination: Aldebaran%s\n  Reward: 5,000 credits + reputation\n  Deadline: 10 stops -- he leaves if not delivered in time\n  Failure: No penalty, ambassador departs", distStr),
			Actions: []string{"Accept passenger", "Decline"},
		}}
	}

	if state == QuestActive {
		aldebaran := findSystem(gs, "Aldebaran")
		if aldebaran >= 0 && gs.CurrentSystemID == aldebaran {
			gs.SetQuestState(QuestJarek, QuestComplete)
			gs.Player.Credits += 5000
			gs.Player.Reputation += 2
			return []QuestEvent{{
				Title:   "Ambassador Delivered!",
				Message: "Ambassador Jarek thanks you. 5,000 credits and increased reputation.",
			}}
		}

		gs.SetQuestProgress(QuestJarek, gs.QuestProgress(QuestJarek)+1)
		if gs.QuestProgress(QuestJarek) > 10 {
			gs.SetQuestState(QuestJarek, QuestUnavailable)
			gs.SetQuestProgress(QuestJarek, 0)
			return []QuestEvent{{
				Title:   "Ambassador Impatient",
				Message: "Ambassador Jarek has lost patience and left your ship at the next port.",
			}}
		}
	}
	return nil
}

func checkGemulon(gs *GameState) []QuestEvent {
	state := gs.QuestState(QuestGemulon)

	if state == QuestUnavailable && gs.Day > 35 && gs.Rand.Intn(100) < 6 {
		_, distStr := systemDistanceStr(gs, "Gemulon")
		gs.SetQuestState(QuestGemulon, QuestAvailable)
		gs.SetQuestProgress(QuestGemulon, gs.Day)
		return []QuestEvent{{
			Title:   "Gemulon Invasion!",
			Message: fmt.Sprintf("Aliens are planning to invade Gemulon! You must warn them within 7 days!\n\n  Destination: Gemulon%s\n  Reward: Fuel Compactor\n  Deadline: 7 days -- system is invaded if you're late\n  Failure: Gemulon falls, quest lost", distStr),
		}}
	}

	if state == QuestAvailable {
		startDay := gs.QuestProgress(QuestGemulon)
		gemulon := findSystem(gs, "Gemulon")
		if gs.Day-startDay > 7 {
			gs.SetQuestState(QuestGemulon, QuestUnavailable)
			gs.SetQuestProgress(QuestGemulon, 0)
			return []QuestEvent{{
				Title:   "Gemulon Invaded",
				Message: "You arrived too late. Gemulon has been invaded.",
			}}
		}
		if gemulon >= 0 && gs.CurrentSystemID == gemulon {
			gs.SetQuestState(QuestGemulon, QuestComplete)
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
	state := gs.QuestState(QuestFehler)

	if state == QuestUnavailable && gs.Day > 40 && gs.Rand.Intn(100) < 5 {
		_, distStr := systemDistanceStr(gs, "Deneb")
		gs.SetQuestState(QuestFehler, QuestAvailable)
		gs.SetQuestProgress(QuestFehler, gs.Day)
		return []QuestEvent{{
			Title:   "Dr. Fehler's Experiment",
			Message: fmt.Sprintf("Dr. Fehler's experiment at Deneb is about to rip a hole in spacetime! Someone must stop it within 5 days!\n\n  Destination: Deneb%s\n  Reward: 10,000 credits + reputation\n  Deadline: 5 days -- spacetime distorted if you're late\n  Failure: Quest lost, no penalty", distStr),
		}}
	}

	if state == QuestAvailable {
		startDay := gs.QuestProgress(QuestFehler)
		deneb := findSystem(gs, "Deneb")

		if gs.Day-startDay > 5 {
			gs.SetQuestState(QuestFehler, QuestUnavailable)
			gs.SetQuestProgress(QuestFehler, 0)
			gs.Quests.FabricRipDays = 25
			gs.SetQuestState(QuestFabricRip, QuestActive)
			return []QuestEvent{{
				Title:   "Experiment Failed!",
				Message: "The experiment tore a hole in spacetime! Random warps may occur during travel for some time.",
			}}
		}
		if deneb >= 0 && gs.CurrentSystemID == deneb {
			gs.SetQuestState(QuestFehler, QuestComplete)
			gs.Player.Credits += 10000
			gs.Player.Reputation += 3
			gs.Quests.HasSingularity = true
			return []QuestEvent{{
				Title:   "Experiment Stopped!",
				Message: "You stopped Dr. Fehler's experiment just in time! 10,000 credits reward.\n\nYou also recovered a Portable Singularity -- use it to warp to any system once!",
			}}
		}
	}
	return nil
}

func checkWild(gs *GameState) []QuestEvent {
	state := gs.QuestState(QuestWild)

	if state == QuestUnavailable && gs.Day > 18 && gs.Rand.Intn(100) < 7 {
		sys := gs.Data.Systems[gs.CurrentSystemID]
		if sys.PoliticalSystem == gamedata.PolAnarchy || sys.PoliticalSystem == gamedata.PolFeudal {
			_, distStr := systemDistanceStr(gs, "Adahn")
			gs.SetQuestState(QuestWild, QuestAvailable)
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
			gs.SetQuestState(QuestWild, QuestComplete)
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
	state := gs.QuestState(QuestReactor)

	if state == QuestUnavailable && gs.Day > 45 && gs.Rand.Intn(100) < 5 {
		dp := &GameDataProvider{Data: gs.Data}
		if gs.Player.FreeCargo(dp) >= 5 {
			_, distStr := systemDistanceStr(gs, "Eridani")
			gs.SetQuestState(QuestReactor, QuestAvailable)
			return []QuestEvent{{
				Title:   "Reactor Delivery",
				Message: fmt.Sprintf("Henry Morgan needs an unstable reactor delivered to Eridani.\n\n  Destination: Eridani%s\n  Reward: Morgan's Laser\n  Cost: 5 cargo bays while carrying\n  Risk: Reactor leaks fuel (-1 per stop)\n  Deadline: None", distStr),
				Actions: []string{"Accept the reactor", "Decline"},
			}}
		}
	}

	if state == QuestActive {
		gs.SetQuestProgress(QuestReactor, gs.QuestProgress(QuestReactor)+1)
		status := gs.QuestProgress(QuestReactor)

		gs.Player.Ship.Fuel -= 1
		if gs.Player.Ship.Fuel < 0 {
			gs.Player.Ship.Fuel = 0
		}

		if status >= 21 {
			gs.SetQuestState(QuestReactor, QuestUnavailable)
			gs.SetQuestProgress(QuestReactor, 0)
			damage := gs.Data.Ships[gs.Player.Ship.TypeID].Hull / 2
			gs.Player.Ship.Hull -= damage
			if gs.Player.Ship.Hull < 1 {
				gs.Player.Ship.Hull = 1
			}
			return []QuestEvent{{
				Title:   "Reactor Meltdown!",
				Message: fmt.Sprintf("The reactor has melted down! Massive damage (%d hull) and the reactor is lost!", damage),
			}}
		}

		eridani := findSystem(gs, "Eridani")
		if eridani >= 0 && gs.CurrentSystemID == eridani {
			gs.SetQuestState(QuestReactor, QuestComplete)
			gs.SetQuestProgress(QuestReactor, 0)
			reward := giveQuestEquipment(gs, "Morgan's Laser")
			return []QuestEvent{{
				Title:   "Reactor Delivered!",
				Message: fmt.Sprintf("Henry Morgan is pleased. %s", reward),
			}}
		}

		if status > 15 {
			return []QuestEvent{{
				Title:   "Reactor Warning",
				Message: fmt.Sprintf("The reactor is becoming unstable! Status: %d/20. Deliver it soon!", status),
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
				gs.SetQuestState(QuestSpaceMonster, QuestComplete)
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
			hasPulse := false
			hasMorgans := false
			for _, wID := range gs.Player.Ship.Weapons {
				name := gs.Data.Equipment[wID].Name
				if name == "Pulse Laser" {
					hasPulse = true
				}
				if name == "Morgan's Laser" {
					hasMorgans = true
				}
			}
			if !hasPulse && !hasMorgans {
				return "Your weapons have no effect on the Scarab's hull! It seems impervious to energy weapons. Perhaps a simpler weapon would work..."
			}
			gs.SetQuestState(QuestScarab, QuestComplete)
			shipDef := gs.Data.Ships[gs.Player.Ship.TypeID]
			gs.Player.Ship.Hull = shipDef.Hull + 20
			return "You destroyed the Scarab! Its hull plating was salvaged -- your hull is permanently reinforced (+20 max hull)."
		}
		return "You leave the Scarab alone."

	case "Alien Artifact":
		if actionIdx == 0 {
			gs.SetQuestState(QuestAlienArtifact, QuestActive)
			return "You take the alien artifact. Mantis ships may now pursue you during travel."
		}
		gs.SetQuestState(QuestAlienArtifact, QuestUnavailable)
		return "You leave the artifact."

	case "Ambassador Jarek":
		if actionIdx == 0 {
			gs.SetQuestState(QuestJarek, QuestActive)
			gs.SetQuestProgress(QuestJarek, 0)
			return "Ambassador Jarek boards your ship. Deliver him to Aldebaran."
		}
		gs.SetQuestState(QuestJarek, QuestUnavailable)
		return "Declined."

	case "Jonathan Wild":
		if actionIdx == 0 {
			gs.SetQuestState(QuestWild, QuestActive)
			return "Jonathan Wild is now aboard. Deliver him to Adahn -- but watch out for police."
		}
		gs.SetQuestState(QuestWild, QuestUnavailable)
		return "You refuse to smuggle a criminal."

	case "Reactor Delivery":
		if actionIdx == 0 {
			dp := &GameDataProvider{Data: gs.Data}
			if gs.Player.FreeCargo(dp) < 5 {
				return "Not enough cargo space for the reactor (need 5 free bays)."
			}
			gs.SetQuestState(QuestReactor, QuestActive)
			gs.SetQuestProgress(QuestReactor, 5)
			return "Reactor loaded (5 cargo bays). It will slowly leak fuel. Deliver to Eridani."
		}
		gs.SetQuestState(QuestReactor, QuestUnavailable)
		return "Declined."
	}
	return ""
}
