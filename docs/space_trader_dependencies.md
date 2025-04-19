# Space Trader Class Hierarchy & Dependency Graph

## Overview
This document outlines the class hierarchy and the chain of dependencies for all core functionality within the Space Trader TUI project, based on the codebase and the design documentation in the /docs folder.

---

## 1. Core Classes and Data Flow

### GameState (game.py)
- **Responsibilities:**
  - Manages the overall game state, including player, galaxy (list of StarSystem), current system, and events.
  - Handles saving/loading game state.
- **Dependencies:**
  - `Player` (player.py)
  - `StarSystem` (world.py)
  - Data files: `systems.json`, `goods.json`, `ships.json`, `equipment.json`, `political_systems.json`, `tech_levels.json`

### Player (player.py)
- **Responsibilities:**
  - Stores player attributes: name, credits, skills, reputation, police record, ship, inventory, mercenaries.
- **Dependencies:**
  - `Ship` (ship.py)
  - `Mercenary` (ship.py)

### Ship (ship.py)
- **Responsibilities:**
  - Represents a spaceship with stats, cargo, equipment, crew, and range.
- **Dependencies:**
  - Equipment (from `equipment.json`)

### Mercenary (ship.py)
- **Responsibilities:**
  - Represents a mercenary crew member with skills and wage.

### StarSystem (world.py)
- **Responsibilities:**
  - Represents a single star system with name, tech level, political system, resources, events, and market.
- **Dependencies:**
  - `Market` (market.py)
  - Data: Loaded from `systems.json`

### Market (market.py)
- **Responsibilities:**
  - Holds available goods and their prices for a given system.
- **Dependencies:**
  - `Good` (market.py)
  - Data: Loaded from `goods.json`, influenced by system attributes and events.

### Good (market.py)
- **Responsibilities:**
  - Represents a tradeable good, with price, legality, tech requirements, etc.

### Encounter (encounter.py)
- **Responsibilities:**
  - Represents random encounters (pirate, trader, police) during travel.
- **Dependencies:**
  - `Player`, `StarSystem`, `Ship` (for combat/trade logic)

---

## 2. UI Layer (ui/screens.py, ui/widgets.py, ui/input.py)
- **Screen Classes:**
  - `MainMenuScreen`, `MarketScreen`, `MapScreen`, `StatusScreen`, `ShipyardScreen`, `TravelScreen`, `EncounterScreen`, `BankScreen`, `PersonnelScreen`
- **Widget Classes:**
  - `MenuWidget`, `InventoryWidget`, `ButtonWidget`
- **Input Handling:**
  - `InputHandler` (ui/input.py)
- **Dependencies:**
  - All depend on the core game state (GameState, Player, StarSystem, etc.) for data display and manipulation.

---

## 3. Data File Dependencies
- systems.json → StarSystem objects
- goods.json → Good objects
- ships.json → Ship creation
- equipment.json → Ship equipment
- political_systems.json, tech_levels.json → StarSystem, Market, Encounter logic

---

## 4. Functional Dependency Chain

1. **GameState** (entry point)
    - loads Player, Galaxy (list of StarSystem), handles save/load
2. **Player**
    - owns Ship, Inventory, Mercenaries
3. **Ship**
    - has Equipment, Crew, Cargo
4. **Galaxy**
    - list of StarSystem objects
5. **StarSystem**
    - has Market, Events, Resources
6. **Market**
    - holds Good objects, prices (from goods.json, affected by system state)
7. **Encounter**
    - interacts with Player, Ship, StarSystem during travel
8. **UI**
    - presents and manipulates all of the above

---

## 5. Diagram (Textual)

```
GameState
  ├── Player
  │     ├── Ship
  │     │     └── Equipment (from equipment.json)
  │     └── Mercenaries
  ├── Galaxy (List[StarSystem])
  │     ├── Market
  │     │     └── Goods (from goods.json)
  │     └── Events, Resources
  └── CurrentSystem

UI Layer
  └── interacts with GameState, Player, StarSystem, Market, etc.

Encounter
  └── interacts with Player, Ship, StarSystem
```

---

## 6. Summary Table

| Class         | Instantiates/Depends On     | Data Source(s)                     |
|---------------|----------------------------|-------------------------------------|
| GameState     | Player, StarSystem         | All .json data files                |
| Player        | Ship, Mercenary            | ships.json, equipment.json          |
| Ship          | Equipment                  | equipment.json                      |
| StarSystem    | Market, Events, Resources  | systems.json, goods.json            |
| Market        | Good                       | goods.json, system state            |
| Encounter     | Player, Ship, StarSystem   | N/A (runtime)                       |
| UI Screens    | GameState, Player, etc.    | N/A (runtime)                       |

---

## 7. Notes
- All data files are loaded at game start and used to instantiate core objects.
- UI is decoupled from logic, but always depends on the current GameState.
- Encounters are generated dynamically during travel.

---

[Back to PRD](space_trader_prd.md)
