# Space Trader Terminal Port — Product Requirements Document (PRD)

## 1. Overview
Space Trader is an open-ended, single-player trading and space adventure game. The player takes on the role of a space trader, traveling between star systems, engaging in commerce, combat, and occasional piracy, with the ultimate goal of acquiring enough wealth to retire on their own moon.

This PRD details the mechanics, systems, and requirements for a terminal-based (TUI) port of Space Trader for Linux.

---

## 2. Core Gameplay Loop
- Player starts with a basic ship, some credits, and a random set of trade goods.
- Player travels between star systems, buying and selling goods for profit.
- Encounters with pirates, police, and other traders occur during travel.
- Player can upgrade their ship, hire mercenaries, and pursue different play styles (trader, bounty hunter, pirate).
- The game ends when the player retires (buys a moon) or is killed.

---

## 3. Major Systems & Mechanics

### 3.1 Trading
- Each system’s market offers goods for sale and buys goods from the player.
- Prices depend on system tech level, government, resources, special events, and random fluctuations.
- Ten trade goods: Water, Furs, Food, Ore, Games, Firearms (illegal), Medicine, Machines, Narcotics (illegal), Robots.
- Illegal goods (Firearms, Narcotics) are subject to police inspection/confiscation.
- Special events (drought, war, plague, boredom, etc.) can dramatically affect prices.
- Cargo capacity and cash limit how much the player can buy.
- Loans can be taken from the bank (10% daily interest, auto-deducted on warp).

### 3.2 Spaceships & Equipment
- Multiple ship types, each with:
  - Hull strength
  - Cargo bays
  - Weapon/shield/gadget slots
  - Crew quarters
  - Max travel range
- Ship upgrades:
  - Weapons (Pulse, Beam, Military lasers)
  - Shields (Energy, Reflective)
  - Gadgets (Extra cargo, Navigation, Targeting, Auto-repair, Cloaking)
  - Escape pod (auto-ejects on destruction)
  - Insurance (refunds value if escape pod used; daily premium, no-claim bonus)
- Ships can be bought/sold; trade-in value includes current cargo/equipment.

### 3.3 Skills & Mercenaries
- Four skills: Pilot, Fighter, Trader, Engineer.
- Player has base skill values; can hire mercenaries to supplement skills.
- Mercenaries are paid daily (on warp); leave if unpaid.
- Skills affect:
  - Piloting: Fleeing/dodging/attacking in combat
  - Fighting: Weapon accuracy
  - Trading: Prices paid/received
  - Engineering: Ship/shield repair, weapon effectiveness

### 3.4 Travel
- Player selects destination from a short-range chart (limited by fuel/tank size).
- Travel steps:
  1. Leave spaceport
  2. Warp (fuel consumed)
  3. Encounter phase (pirates, traders, police)
  4. Dock at destination
- Wormholes: allow instant travel between distant systems for a fee (no fuel used).

### 3.5 Encounters
- Random encounters during travel:
  - Traders: May offer deals, can be attacked or ignored
  - Pirates: Usually attack, may flee if player is strong
  - Police: May inspect cargo, can be bribed (except in incorruptible governments), may attack criminals
- Combat: Options to attack, flee, surrender, or attempt bribes (contextual)
- Outcomes: Surrender, destruction (with/without escape pod), loot, fines, prison

### 3.6 Economy
- Prices for goods, ships, and equipment fluctuate by:
  - Tech level
  - Political system
  - Resources
  - Special situations/events
- Each system has its own market and characteristics.

---

## 4. World Generation
- Galaxy composed of multiple star systems, each with:
  - Name
  - Tech level (Pre-agricultural to Hi-tech)
  - Political system (Anarchy, Democracy, Theocracy, etc.)
  - Special resources/events
  - Market (goods, prices)
- System attributes influence encounters, prices, and available goods/equipment.

---

## 5. User Interface (TUI)
- All interactions via keyboard (no mouse required)
- Main screens:
  - Status (ship, crew, cargo, credits, insurance, police record)
  - Star system map/short-range chart
  - Market (buy/sell goods)
  - Shipyard (buy/sell ships, repairs, upgrades)
  - Personnel (hire/fire mercenaries)
  - Bank (loans, insurance)
  - Travel (select destination, refuel)
  - Encounter (combat, trade, bribe, flee, surrender)
- Clear prompts and feedback for all actions
- Save/load game support

---

## 6. Game Progression & End States
- Player can pursue:
  - Honest trading
  - Bounty hunting
  - Piracy
- Reputation and police record affect encounters and opportunities
- Game ends if:
  - Player retires (buys a moon)
  - Player is killed (no escape pod)

---

## 7. Special Events & Tips
- Dynamic events (war, famine, boredom, etc.) affect system markets
- Special assignments/quests may be offered (with risk/reward)
- Tips:
  - Always keep some cargo (even cheap) to avoid cash extortion by pirates
  - Richer players attract stronger pirates
  - Police/criminal response escalates with reputation
  - Stockpiling/trading between two systems only works short-term

---

## 8. Technical Requirements
- Runs on Linux terminal (ncurses or similar TUI library)
- Written in a portable language (Python, C, or Rust recommended)
- Modular codebase for future extension
- Configurable for different terminal sizes

---

## 9. Non-Goals
- No graphical or mouse-driven interface
- No multiplayer support
- No persistent online features

---

## 10. References
- [Original Space Trader Documentation](https://github.com/SpaceTraderGame/SpaceTrader-Windows/blob/master/Space%20Trader%20for%20Windows%20Documentation.htm)
- [Space Trader for Windows Source Code](https://github.com/SpaceTraderGame/SpaceTrader-Windows)

---

## Appendix: Detailed System Specifications

For exhaustive details on each major system or mechanic, see these sub-documents:

- [Trading System](space_trader_trading.md): Economic model, trade goods, pricing, events, and legality.
- [Spaceships & Equipment](space_trader_ships.md): Ship types, upgrades, equipment, repairs, and insurance.
- [Skills & Mercenaries](space_trader_skills.md): Player/mercenary skills, hiring, payment, and effects.
- [Travel & Encounters](space_trader_travel.md): Travel mechanics, encounters, combat, and police/reputation.
- [World Generation & Economy](space_trader_world.md): Star system attributes, tech levels, political systems, special events, and market dynamics.

Each sub-document contains implementation-level details to faithfully reproduce the original game's mechanics in the terminal port.

This PRD will be updated as development progresses and requirements are refined.
