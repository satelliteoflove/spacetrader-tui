# Space Trader: Terminal Edition

A terminal-based port of the classic PalmOS game **Space Trader** by Pieter Spronck, built with Go and [Bubbletea](https://github.com/charmbracelet/bubbletea).

Trade goods between star systems, upgrade your ship, dodge pirates, and earn enough credits to buy your own moon and retire.

## Features

- Full trading economy with 10 goods across 78 star systems
- Profit/loss tracking on cargo with cost basis per good
- Ship upgrades, weapons, shields, and gadgets via the shipyard
- Encounters with pirates, police, and traders during warp travel
- Threat assessment before combat (scanner readout of enemy strength)
- Banking system with loans and interest
- Crew hiring and management
- Galactic chart with free-moving grid cursor and visual star map
- Sortable/filterable system lists with color-coded specialties
- Bookmark systems from news or charts, searchable via /
- Recent News screen tracking event headlines with age
- Trader's Guide with in-game reference for tech, government, and specialties
- 10 unique quest lines (Dragonfly, Space Monster, Alien Artifact, and more)
- Wormhole travel network
- Persistent status bar showing credits, cargo, hull, fuel, and day
- Confirmation prompts for travel, ship purchases, and crew changes
- Difficulty affects skill point pool, pirate strength, and score multiplier
- Save/load game support
- Runs great on a Raspberry Pi Zero W

## Requirements

- Go 1.21 or later
- A terminal emulator with 80+ column support

## Build and Run

```sh
# Build and run locally
make run

# Build only
make build

# Cross-compile for Raspberry Pi Zero W
make build-pi
# Then copy spacetrader-arm to your Pi and run it
```

Or directly with Go:

```sh
go run .
```

## Controls

| Key           | Action                                      |
|---------------|---------------------------------------------|
| j / k         | Navigate up/down                            |
| h / l         | Navigate left/right (galactic map, skills)  |
| Arrow keys    | Navigate (all directions on galactic map)   |
| Enter         | Select / confirm / travel                   |
| Esc           | Back / cancel                               |
| 1-5           | Sort by column (charts)                     |
| /             | Filter systems (charts) / search (map)      |
| b             | Buy (market) / toggle bookmark (charts/news)|
| s             | Sell (market) / save game (system hub)      |
| r             | Refuel (short-range chart)                  |
| w             | Use wormhole                                |
| L             | Switch to list view (galactic map)          |
| Ctrl+C        | Quit                                        |

## Project Structure

```
main.go                  Entry point
tui/
  app.go                 Top-level Bubbletea model and routing
  screens/               One file per game screen
    shared.go            Shared styles, key bindings, helpers
    system_table.go      Sort/filter logic for system lists
    system.go            System hub (main menu per system)
    market.go            Buy/sell goods with profit tracking
    chart.go             Short-range chart (travel)
    galactic_chart.go    Galactic map with grid cursor and system list
    shipyard_screen.go   Ships, equipment, repairs
    bank.go              Loans
    personnel.go         Crew management
    status.go            Commander status
    encounter.go         Combat and encounters with threat assessment
    quest_event.go       Quest narrative events
    newgame.go           Character creation (difficulty, skills, descriptions)
    guide.go             Trader's Guide reference screen
    news.go              Recent News with headline tracking
    gameover.go          Retirement/death screen
    save.go              Save game
internal/
  game/                  Core game state, quests, travel
  gamedata/              Type definitions and enums
  data/                  JSON data loader
  economy/               Banking, loans, scoring
  market/                Trading and price calculations
  shipyard/              Ship purchases and repairs
  encounter/             Combat resolution
  travel/                Travel and distance calculations
  formula/               Shared game formulas
data/
  *.json                 Game data files (systems, ships, goods, etc.)
```

## Credits

Based on [Space Trader](https://www.spronck.net/spacetrader/) by Pieter Spronck, originally released for PalmOS in 2000.

## License

MIT
