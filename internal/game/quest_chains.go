package game

import (
	"fmt"

	"github.com/the4ofus/spacetrader-tui/internal/formula"
	"github.com/the4ofus/spacetrader-tui/internal/gamedata"
)

type QuestCombatResult struct {
	Log    []CombatLogLine
	Result string
}

type QuestActionResult struct {
	Message string
	Combat  *QuestCombatResult
}

var JarekSkills = [formula.NumSkills]int{3, 2, 10, 4}
var WildSkills = [formula.NumSkills]int{7, 10, 2, 5}

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
	hops := gs.HopsToSystem(idx)
	if hops > 0 {
		return idx, fmt.Sprintf(" (%.1f parsecs, ~%d hops)", dist, hops)
	}
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

func tryGiveQuestEquipment(gs *GameState, equipName string) (string, bool) {
	eqID := findEquipByName(gs, equipName)
	if eqID < 0 {
		return "", false
	}
	eq := gs.Data.Equipment[eqID]
	shipDef := gs.Data.Ships[gs.Player.Ship.TypeID]

	switch eq.Category {
	case gamedata.EquipWeapon:
		if len(gs.Player.Ship.Weapons) < shipDef.WeaponSlots {
			gs.Player.Ship.Weapons = append(gs.Player.Ship.Weapons, eqID)
			return fmt.Sprintf("Received %s!", equipName), true
		}
		return fmt.Sprintf("No free weapon slot for %s. Free a slot and return here to claim it.", equipName), false
	case gamedata.EquipShield:
		if len(gs.Player.Ship.Shields) < shipDef.ShieldSlots {
			gs.Player.Ship.Shields = append(gs.Player.Ship.Shields, eqID)
			return fmt.Sprintf("Received %s!", equipName), true
		}
		return fmt.Sprintf("No free shield slot for %s. Free a slot and return here to claim it.", equipName), false
	case gamedata.EquipGadget:
		if len(gs.Player.Ship.Gadgets) < shipDef.GadgetSlots {
			gs.Player.Ship.Gadgets = append(gs.Player.Ship.Gadgets, eqID)
			return fmt.Sprintf("Received %s!", equipName), true
		}
		return fmt.Sprintf("No free gadget slot for %s. Free a slot and return here to claim it.", equipName), false
	}
	return "", false
}

func addPendingReward(gs *GameState, questID QuestID, equipment string, systemIdx int) {
	for _, pr := range gs.Quests.PendingRewards {
		if pr.QuestID == questID {
			return
		}
	}
	gs.Quests.PendingRewards = append(gs.Quests.PendingRewards, PendingReward{
		QuestID:   questID,
		Equipment: equipment,
		SystemIdx: systemIdx,
	})
}

func CheckPendingRewards(gs *GameState) []QuestEvent {
	var events []QuestEvent
	var remaining []PendingReward
	for _, pr := range gs.Quests.PendingRewards {
		if gs.CurrentSystemID == pr.SystemIdx {
			msg, installed := tryGiveQuestEquipment(gs, pr.Equipment)
			if installed {
				events = append(events, QuestEvent{
					Title:   "Equipment Installed!",
					Message: msg,
				})
			} else {
				events = append(events, QuestEvent{
					Title:   "Equipment Available",
					Message: msg,
				})
				remaining = append(remaining, pr)
			}
		} else {
			remaining = append(remaining, pr)
		}
	}
	gs.Quests.PendingRewards = remaining
	return events
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
			if progress == len(dragonflyPath)-1 {
				dragonflyHull := gs.Quests.DragonflyHull
				if dragonflyHull == 0 {
					dragonflyHull = DragonflyMaxHull
				}
				return []QuestEvent{{
					Title:   "Dragonfly Cornered!",
					Message: fmt.Sprintf("The Dragonfly is cornered and can't escape! Its experimental Lightning Shield crackles with energy.\n  Dragonfly hull: %d/%d", dragonflyHull, DragonflyMaxHull),
					Actions: []string{"Attack!", "Back off"},
				}}
			}
			progress++
			gs.SetQuestProgress(QuestDragonfly, progress)
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

	acamar := findSystem(gs, "Acamar")

	if state == QuestUnavailable && gs.Day > 15 && gs.Rand.Intn(100) < 8 {
		_, distStr := systemDistanceStr(gs, "Acamar")
		if distStr != "" {
			distStr = "\n  Location: Acamar" + distStr
		}
		gs.SetQuestState(QuestSpaceMonster, QuestAvailable)
		events := []QuestEvent{{
			Title:   "Space Monster",
			Message: fmt.Sprintf("A terrifying Space Monster is attacking ships near Acamar! Bounty offered for its destruction.\n%s\n  Reward: 10,000 credits + reputation\n  Risk: Combat -- strength depends on fighter skill and weapons\n  Deadline: None", distStr),
		}}
		if gs.CurrentSystemID == acamar {
			monsterHull := gs.Quests.MonsterHull
			if monsterHull == 0 {
				monsterHull = MonsterMaxHull
			}
			events = append(events, QuestEvent{
				Title:   "Space Monster!",
				Message: fmt.Sprintf("The Space Monster attacks! Monster hull: %d/%d. It regenerates 5%% hull per day -- don't delay!", monsterHull, MonsterMaxHull),
				Actions: []string{"Fight!", "Flee"},
			})
		}
		return events
	}

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
			Message: "Captain Renwick's Scarab has been spotted near a wormhole exit. He developed an organic hull that cannot be damaged except by Pulse lasers.\n\n  Location: Near wormhole exits (chance encounter)\n  Reward: Salvaged hull plating (+50 max hull)\n  Deadline: None",
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

	if state == QuestUnavailable && gs.Day > 12 && gs.Rand.Intn(100) < 10 && FreeCrewQuarters(gs) > 0 {
		devIdx, distStr := systemDistanceStr(gs, "Devidia")
		if devIdx >= 0 && gs.HopsToSystem(devIdx) > 10 {
			return nil
		}
		gs.SetQuestState(QuestJarek, QuestAvailable)
		return []QuestEvent{{
			Title:   "Ambassador Jarek",
			Message: fmt.Sprintf("Ambassador Jarek needs transport to Devidia. A recent political crisis has forced him to flee.\n\n  Destination: Devidia%s\n  Reward: Haggling Computer (+1 Trader skill)\n  Deadline: 10 stops -- he leaves if not delivered in time\n  Note: Uses one crew quarter", distStr),
			Actions: []string{"Accept passenger", "Decline"},
		}}
	}

	if state == QuestActive {
		devidia := findSystem(gs, "Devidia")
		if devidia >= 0 && gs.CurrentSystemID == devidia {
			gs.SetQuestState(QuestJarek, QuestComplete)
			RemoveQuestCrew(gs, "Jarek")
			return []QuestEvent{{
				Title:   "Ambassador Delivered!",
				Message: "Ambassador Jarek is very grateful. As a reward, he gives you an experimental handheld haggling computer, which gives you larger discounts on purchases.",
			}}
		}

		gs.SetQuestProgress(QuestJarek, gs.QuestProgress(QuestJarek)+1)
		progress := gs.QuestProgress(QuestJarek)
		if progress == 5 {
			return []QuestEvent{{
				Title:   "Jarek Concerned",
				Message: "Ambassador Jarek is wondering why the journey is taking so long.",
			}}
		}
		if progress == 9 {
			for i, m := range gs.Player.Crew {
				if m.Name == "Jarek" && m.IsQuest {
					gs.Player.Crew[i].Skills = [formula.NumSkills]int{0, 0, 0, 0}
					break
				}
			}
			return []QuestEvent{{
				Title:   "Jarek Impatient",
				Message: "Ambassador Jarek is no longer of much help in negotiating trades.",
			}}
		}
		if progress > 10 {
			gs.SetQuestState(QuestJarek, QuestUnavailable)
			gs.SetQuestProgress(QuestJarek, 0)
			RemoveQuestCrew(gs, "Jarek")
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
		gemIdx, distStr := systemDistanceStr(gs, "Gemulon")
		if gemIdx >= 0 && gs.HopsToSystem(gemIdx) >= 7 {
			return nil
		}
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
			if gemulon >= 0 {
				gs.Data.Systems[gemulon].TechLevel = gamedata.TechPreAgricultural
				gs.Data.Systems[gemulon].PoliticalSystem = gamedata.PolAnarchy
			}
			return []QuestEvent{{
				Title:   "Gemulon Invaded",
				Message: "You arrived too late. Gemulon has been invaded. The system has fallen to anarchy.",
			}}
		}
		if gemulon >= 0 && gs.CurrentSystemID == gemulon {
			reward, installed := tryGiveQuestEquipment(gs, "Fuel Compactor")
			if installed {
				gs.SetQuestState(QuestGemulon, QuestComplete)
			} else {
				addPendingReward(gs, QuestGemulon, "Fuel Compactor", gemulon)
				gs.SetQuestState(QuestGemulon, QuestComplete)
			}
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
		denIdx, distStr := systemDistanceStr(gs, "Deneb")
		if denIdx >= 0 && gs.HopsToSystem(denIdx) >= 5 {
			return nil
		}
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

	if state == QuestUnavailable && gs.Day > 18 && gs.Rand.Intn(100) < 7 && FreeCrewQuarters(gs) > 0 {
		sys := gs.Data.Systems[gs.CurrentSystemID]
		if sys.PoliticalSystem == gamedata.PolAnarchy || sys.PoliticalSystem == gamedata.PolFeudal {
			_, distStr := systemDistanceStr(gs, "Kravat")
			gs.SetQuestState(QuestWild, QuestAvailable)
			return []QuestEvent{{
				Title:   "Jonathan Wild",
				Message: fmt.Sprintf("The notorious criminal Jonathan Wild wants passage to Kravat.\n\n  Destination: Kravat%s\n  Reward: Police record cleaned + Zeethibal (free mercenary)\n  Risk: Increased police encounters\n  Requires: Beam Laser, free crew quarter\n  Deadline: None", distStr),
				Actions: []string{"Take him aboard", "Refuse"},
			}}
		}
	}

	if state == QuestActive {
		kravat := findSystem(gs, "Kravat")
		if kravat >= 0 && gs.CurrentSystemID == kravat {
			gs.SetQuestState(QuestWild, QuestComplete)
			RemoveQuestCrew(gs, "Wild")
			gs.Player.PoliceRecord = 0
			CreateZeethibal(gs)
			return []QuestEvent{{
				Title:   "Wild Delivered",
				Message: "Jonathan Wild is most grateful. As a reward, he has one of his Cyber Criminals hack into the Police Database and clean up your record. He also offers you the opportunity to take his talented nephew Zeethibal along as a mercenary with no pay.\n\nVisit the Personnel screen here at Kravat to hire Zeethibal.",
			}}
		}

		gs.SetQuestProgress(QuestWild, gs.QuestProgress(QuestWild)+1)
		progress := gs.QuestProgress(QuestWild)
		if progress == 5 {
			return []QuestEvent{{
				Title:   "Wild Concerned",
				Message: "Jonathan Wild is wondering why the journey is taking so long.",
			}}
		}
		if progress == 9 {
			for i, m := range gs.Player.Crew {
				if m.Name == "Wild" && m.IsQuest {
					gs.Player.Crew[i].Skills = [formula.NumSkills]int{0, 0, 0, 0}
					break
				}
			}
			return []QuestEvent{{
				Title:   "Wild Impatient",
				Message: "Jonathan Wild is getting impatient, and will no longer aid your crew along the way.",
			}}
		}
		if progress > 10 {
			gs.SetQuestState(QuestWild, QuestUnavailable)
			gs.SetQuestProgress(QuestWild, 0)
			RemoveQuestCrew(gs, "Wild")
			return []QuestEvent{{
				Title:   "Wild Leaves",
				Message: "Jonathan Wild has left your ship and gone into hiding.",
			}}
		}
	}
	return nil
}

func ReactorOnBoard(gs *GameState) bool {
	status := gs.QuestProgress(QuestReactor)
	return gs.QuestState(QuestReactor) == QuestActive && status >= ReactorStatusFuelOk && status < ReactorStatusDate
}

func ReactorCargoBays(gs *GameState) int {
	if !ReactorOnBoard(gs) {
		return 0
	}
	status := gs.QuestProgress(QuestReactor)
	fuelBays := ReactorFuelBays - (status-1)/2
	if fuelBays < 0 {
		fuelBays = 0
	}
	return ReactorBays + fuelBays
}

func checkReactor(gs *GameState) []QuestEvent {
	state := gs.QuestState(QuestReactor)
	status := gs.QuestProgress(QuestReactor)

	if state == QuestUnavailable && status == ReactorStatusNotStarted &&
		gs.Day > 45 && gs.Player.PoliceRecord < -5 && gs.Player.Reputation >= 40 {
		if gs.Rand.Intn(100) < 5 {
			nixIdx, distStr := systemDistanceStr(gs, "Nix")
			if nixIdx >= 0 && gs.HopsToSystem(nixIdx) > 20 {
				return nil
			}
			gs.SetQuestState(QuestReactor, QuestAvailable)
			return []QuestEvent{{
				Title:   "Reactor Delivery",
				Message: fmt.Sprintf("Galactic criminal Henry Morgan wants an illegal ion reactor delivered to Nix. It's very dangerous!\n\n  Destination: Nix%s\n  Reward: Morgan's Laser (most powerful weapon)\n  Cost: 15 cargo bays (5 reactor + 10 fuel, fuel frees over time)\n  Risk: Fuel consumed, meltdown if too slow (19 warps)\n  Warning: Reactor is illegal cargo!", distStr),
				Actions: []string{"Accept the reactor", "Decline"},
			}}
		}
	}

	if state == QuestActive && status >= ReactorStatusFuelOk && status < ReactorStatusDate {
		nix := findSystem(gs, "Nix")
		if nix >= 0 && gs.CurrentSystemID == nix {
			reward, installed := tryGiveQuestEquipment(gs, "Morgan's Laser")
			if installed {
				gs.SetQuestProgress(QuestReactor, ReactorStatusDone)
				gs.SetQuestState(QuestReactor, QuestComplete)
			} else {
				gs.SetQuestProgress(QuestReactor, ReactorStatusDelivered)
				addPendingReward(gs, QuestReactor, "Morgan's Laser", nix)
			}
			return []QuestEvent{{
				Title:   "Reactor Delivered!",
				Message: fmt.Sprintf("Henry Morgan takes delivery with great glee. His men immediately stabilize the fuel system. %s", reward),
			}}
		}

		var events []QuestEvent
		switch status {
		case ReactorStatusFuelOk + 1:
			events = append(events, QuestEvent{
				Title:   "Reactor Warning",
				Message: "The Ion Reactor has begun to consume fuel rapidly. In a single day, it burned nearly half a bay of fuel!",
			})
		case ReactorStatusDate - 4:
			events = append(events, QuestEvent{
				Title:   "Reactor Warning",
				Message: "The Ion Reactor is emitting a shrill whine and shaking. The display indicates fuel starvation.",
			})
		case ReactorStatusDate - 2:
			events = append(events, QuestEvent{
				Title:   "Reactor Critical!",
				Message: "The Ion Reactor is smoking and making loud noises. The core is close to melting temperature!",
			})
		}
		return events
	}

	if state == QuestActive && status >= ReactorStatusDate {
		if gs.Player.HasEscapePod {
			gs.SetQuestProgress(QuestReactor, ReactorStatusNotStarted)
			gs.SetQuestState(QuestReactor, QuestUnavailable)
			ClearCrewAndResetQuests(gs)
			gs.Player.Ship = NewStartingShip(gs.Data)
			gs.Player.Cargo = [10]int{}
			return []QuestEvent{{
				Title:   "Reactor Meltdown!",
				Message: "The reactor explodes! Your escape pod saves you, but your ship and all cargo are lost. You find yourself in a new Flea.",
			}}
		}
		gs.EndStatus = StatusDead
		return []QuestEvent{{
			Title:   "Reactor Meltdown!",
			Message: "The reactor explodes into a huge radioactive fireball! Without an escape pod, you perish in the explosion.",
		}}
	}

	if state == QuestActive && status == ReactorStatusDelivered {
		nix := findSystem(gs, "Nix")
		if nix >= 0 && gs.CurrentSystemID == nix {
			reward, installed := tryGiveQuestEquipment(gs, "Morgan's Laser")
			if installed {
				gs.SetQuestProgress(QuestReactor, ReactorStatusDone)
				gs.SetQuestState(QuestReactor, QuestComplete)
				return []QuestEvent{{
					Title:   "Morgan's Laser",
					Message: reward,
				}}
			}
			return []QuestEvent{{
				Title:   "Morgan's Laser",
				Message: reward,
			}}
		}
	}

	return nil
}

func resolveQuestChainAction(gs *GameState, title string, actionIdx int) QuestActionResult {
	switch title {
	case "Dragonfly Cornered!":
		if actionIdx == 0 {
			return QuestActionResult{Combat: resolveDragonflyCombat(gs)}
		}
		return QuestActionResult{Message: "You back off. The Dragonfly remains cornered here."}

	case "Space Monster!":
		if actionIdx == 0 {
			return QuestActionResult{Combat: resolveMonsterCombat(gs)}
		}
		return QuestActionResult{Message: "You flee from the Space Monster."}

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
				return QuestActionResult{Message: "Your weapons bounce off the Scarab's organic hull! Only Pulse lasers can penetrate it."}
			}
			gs.SetQuestState(QuestScarab, QuestComplete)
			gs.Player.Ship.HullUpgraded = true
			shipDef := gs.Data.Ships[gs.Player.Ship.TypeID]
			gs.Player.Ship.Hull = shipDef.Hull + ScarabHullBonus
			return QuestActionResult{Message: fmt.Sprintf("You destroyed the Scarab! Its hull plating was salvaged -- your hull is permanently reinforced (+%d max hull).", ScarabHullBonus)}
		}
		return QuestActionResult{Message: "You leave the Scarab alone."}

	case "Alien Artifact":
		if actionIdx == 0 {
			gs.SetQuestState(QuestAlienArtifact, QuestActive)
			return QuestActionResult{Message: "You take the alien artifact. Mantis ships may now pursue you during travel."}
		}
		gs.SetQuestState(QuestAlienArtifact, QuestUnavailable)
		return QuestActionResult{Message: "You leave the artifact."}

	case "Ambassador Jarek":
		if actionIdx == 0 {
			if FreeCrewQuarters(gs) <= 0 {
				return QuestActionResult{Message: "You don't have any crew quarters available for a passenger."}
			}
			AddQuestCrew(gs, "Jarek", JarekSkills)
			gs.SetQuestState(QuestJarek, QuestActive)
			gs.SetQuestProgress(QuestJarek, 0)
			return QuestActionResult{Message: "Ambassador Jarek boards your ship. Deliver him to Devidia."}
		}
		gs.SetQuestState(QuestJarek, QuestUnavailable)
		return QuestActionResult{Message: "Declined."}

	case "Jonathan Wild":
		if actionIdx == 0 {
			if FreeCrewQuarters(gs) <= 0 {
				return QuestActionResult{Message: "You don't have any crew quarters available for a passenger."}
			}
			hasBeamLaser := false
			for _, wID := range gs.Player.Ship.Weapons {
				name := gs.Data.Equipment[wID].Name
				if name == "Beam Laser" || name == "Military Laser" || name == "Morgan's Laser" {
					hasBeamLaser = true
					break
				}
			}
			if !hasBeamLaser {
				return QuestActionResult{Message: "Jonathan Wild isn't willing to go with you if you're not armed with at least a Beam Laser."}
			}
			if ReactorOnBoard(gs) {
				return QuestActionResult{Message: "Jonathan Wild doesn't like the looks of that Ion Reactor. He thinks it's too dangerous, and won't get on board."}
			}
			AddQuestCrew(gs, "Wild", WildSkills)
			gs.SetQuestState(QuestWild, QuestActive)
			return QuestActionResult{Message: "Jonathan Wild is now aboard. Deliver him to Kravat -- but watch out for police."}
		}
		gs.SetQuestState(QuestWild, QuestUnavailable)
		return QuestActionResult{Message: "You refuse to smuggle a criminal."}

	case "Reactor Delivery":
		if actionIdx == 0 {
			dp := &GameDataProvider{Data: gs.Data}
			if gs.Player.FreeCargo(dp) < ReactorTotalBays {
				return QuestActionResult{Message: fmt.Sprintf("Not enough cargo space for the reactor (need %d free bays).", ReactorTotalBays)}
			}
			wildWasActive := gs.QuestState(QuestWild) == QuestActive
			if wildWasActive {
				RemoveQuestCrew(gs, "Wild")
				gs.SetQuestState(QuestWild, QuestUnavailable)
				gs.SetQuestProgress(QuestWild, 0)
			}
			gs.SetQuestState(QuestReactor, QuestActive)
			gs.SetQuestProgress(QuestReactor, ReactorStatusFuelOk)
			msg := fmt.Sprintf("Reactor loaded. %d bays contain the reactor, %d bays contain enriched fuel. Deliver to Nix before it melts down!", ReactorBays, ReactorFuelBays)
			if wildWasActive {
				msg += "\n\nJonathan Wild refuses to stay aboard with the reactor and has departed."
			}
			return QuestActionResult{Message: msg}
		}
		gs.SetQuestState(QuestReactor, QuestUnavailable)
		return QuestActionResult{Message: "Declined."}
	}
	return QuestActionResult{}
}

func resolveDragonflyCombat(gs *GameState) *QuestCombatResult {
	if gs.Quests.DragonflyHull <= 0 {
		gs.Quests.DragonflyHull = DragonflyMaxHull
	}

	fighterSkill := EffectivePlayerSkill(gs, formula.SkillFighter)
	engineerSkill := EffectivePlayerSkill(gs, formula.SkillEngineer)

	playerWeaponPower := 0
	for _, w := range gs.Player.Ship.Weapons {
		playerWeaponPower += gs.Data.Equipment[w].Power
	}
	if playerWeaponPower == 0 {
		return &QuestCombatResult{Result: "You have no weapons! The Dragonfly's shields hold easily."}
	}

	dfWeapon := 20
	dfFighter := 6 + int(gs.Difficulty)
	dfPilot := 10 + int(gs.Difficulty)
	dfEngineer := 3 + int(gs.Difficulty)
	shipDef := gs.Data.Ships[gs.Player.Ship.TypeID]

	var lines []CombatLogLine
	maxRounds := 8
	for round := 0; round < maxRounds; round++ {
		playerHit := gs.Rand.Intn(fighterSkill+3) >= gs.Rand.Intn(dfPilot/2+5)
		if playerHit {
			dmg := 1 + gs.Rand.Intn(playerWeaponPower*(100+2*engineerSkill)/100 + 1)
			gs.Quests.DragonflyHull -= dmg
			lines = append(lines, CombatLogLine{
				Attacker: "You", Hit: true, Damage: dmg, HullDamage: dmg, IsPlayer: true,
			})
		} else {
			lines = append(lines, CombatLogLine{
				Attacker: "You", Hit: false, IsPlayer: true,
			})
		}

		if gs.Quests.DragonflyHull <= 0 {
			gs.Quests.DragonflyHull = 0
			lastSys := findSystem(gs, dragonflyPath[len(dragonflyPath)-1])
			reward, installed := tryGiveQuestEquipment(gs, "Lightning Shield")
			if installed {
				gs.SetQuestState(QuestDragonfly, QuestComplete)
			} else {
				addPendingReward(gs, QuestDragonfly, "Lightning Shield", lastSys)
				gs.SetQuestState(QuestDragonfly, QuestComplete)
			}
			gs.SetQuestProgress(QuestDragonfly, len(dragonflyPath))
			gs.Player.Reputation += 3
			return &QuestCombatResult{
				Log:    lines,
				Result: fmt.Sprintf("The Dragonfly explodes! %s", reward),
			}
		}

		dfHit := gs.Rand.Intn(dfFighter+3) >= gs.Rand.Intn(EffectivePlayerSkill(gs, formula.SkillPilot)/2+5)
		if dfHit {
			dmg := 1 + gs.Rand.Intn(dfWeapon*(100+2*dfEngineer)/100 + 1)
			dmg -= gs.Rand.Intn(max(1, EffectivePlayerSkill(gs, formula.SkillPilot)))
			if dmg < 1 {
				dmg = 1
			}
			gs.Player.Ship.Hull -= dmg
			lines = append(lines, CombatLogLine{
				Attacker: "Dragonfly", Hit: true, Damage: dmg, HullDamage: dmg, IsPlayer: false,
			})
		} else {
			lines = append(lines, CombatLogLine{
				Attacker: "Dragonfly", Hit: false, IsPlayer: false,
			})
		}

		if gs.Player.Ship.Hull <= 0 {
			gs.Player.Ship.Hull = 1
			return &QuestCombatResult{
				Log:    lines,
				Result: fmt.Sprintf("You barely escape! Dragonfly hull: %d/%d.", gs.Quests.DragonflyHull, DragonflyMaxHull),
			}
		}
	}

	maxHull := shipDef.Hull
	if gs.Player.Ship.HullUpgraded {
		maxHull += ScarabHullBonus
	}
	return &QuestCombatResult{
		Log: lines,
		Result: fmt.Sprintf("The battle is inconclusive. You disengage.\nYour hull: %d/%d  |  Dragonfly hull: %d/%d",
			gs.Player.Ship.Hull, maxHull, gs.Quests.DragonflyHull, DragonflyMaxHull),
	}
}

func resolveMonsterCombat(gs *GameState) *QuestCombatResult {
	if gs.Quests.MonsterHull <= 0 {
		gs.Quests.MonsterHull = MonsterMaxHull
	}

	fighterSkill := EffectivePlayerSkill(gs, formula.SkillFighter)
	engineerSkill := EffectivePlayerSkill(gs, formula.SkillEngineer)

	playerWeaponPower := 0
	for _, w := range gs.Player.Ship.Weapons {
		playerWeaponPower += gs.Data.Equipment[w].Power
	}
	if playerWeaponPower == 0 {
		return &QuestCombatResult{Result: "You have no weapons! The Space Monster drives you away."}
	}

	monsterWeapon := 35
	monsterFighter := 8 + int(gs.Difficulty)
	monsterPilot := 8 + int(gs.Difficulty)
	monsterEngineer := 1 + int(gs.Difficulty)
	shipDef := gs.Data.Ships[gs.Player.Ship.TypeID]

	var lines []CombatLogLine
	maxRounds := 8
	for round := 0; round < maxRounds; round++ {
		playerHit := gs.Rand.Intn(fighterSkill+3) >= gs.Rand.Intn(monsterPilot/2+5)
		if playerHit {
			dmg := 1 + gs.Rand.Intn(playerWeaponPower*(100+2*engineerSkill)/100 + 1)
			gs.Quests.MonsterHull -= dmg
			lines = append(lines, CombatLogLine{
				Attacker: "You", Hit: true, Damage: dmg, HullDamage: dmg, IsPlayer: true,
			})
		} else {
			lines = append(lines, CombatLogLine{
				Attacker: "You", Hit: false, IsPlayer: true,
			})
		}

		if gs.Quests.MonsterHull <= 0 {
			gs.Quests.MonsterHull = 0
			gs.SetQuestState(QuestSpaceMonster, QuestComplete)
			gs.Player.Credits += 10000
			gs.Player.Reputation += 5
			return &QuestCombatResult{
				Log:    lines,
				Result: "The Space Monster is destroyed! 10,000 credits bounty and fame across the galaxy!",
			}
		}

		monsterHit := gs.Rand.Intn(monsterFighter+3) >= gs.Rand.Intn(EffectivePlayerSkill(gs, formula.SkillPilot)/2+5)
		if monsterHit {
			dmg := 1 + gs.Rand.Intn(monsterWeapon*(100+2*monsterEngineer)/100 + 1)
			dmg -= gs.Rand.Intn(max(1, EffectivePlayerSkill(gs, formula.SkillPilot)))
			if dmg < 1 {
				dmg = 1
			}
			gs.Player.Ship.Hull -= dmg
			lines = append(lines, CombatLogLine{
				Attacker: "Monster", Hit: true, Damage: dmg, HullDamage: dmg, IsPlayer: false,
			})
		} else {
			lines = append(lines, CombatLogLine{
				Attacker: "Monster", Hit: false, IsPlayer: false,
			})
		}

		if gs.Player.Ship.Hull <= 0 {
			gs.Player.Ship.Hull = 1
			return &QuestCombatResult{
				Log:    lines,
				Result: fmt.Sprintf("You barely escape with your ship intact! Monster hull: %d/%d.", gs.Quests.MonsterHull, MonsterMaxHull),
			}
		}
	}

	maxHull := shipDef.Hull
	if gs.Player.Ship.HullUpgraded {
		maxHull += ScarabHullBonus
	}
	return &QuestCombatResult{
		Log: lines,
		Result: fmt.Sprintf("The battle is inconclusive. You disengage.\nYour hull: %d/%d  |  Monster hull: %d/%d",
			gs.Player.Ship.Hull, maxHull, gs.Quests.MonsterHull, MonsterMaxHull),
	}
}
