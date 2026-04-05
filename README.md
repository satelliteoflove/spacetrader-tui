# Space Trader: Terminal Edition

A terminal-based port of the classic PalmOS game **Space Trader** by Pieter Spronck, built with Go and [Bubbletea](https://github.com/charmbracelet/bubbletea).

Trade goods between star systems, upgrade your ship, dodge pirates, and earn enough credits to buy your own moon and retire.

## Features

- Full trading economy with 10 goods across 78 star systems
- Ship upgrades, weapons, shields, and gadgets via the shipyard
- Encounters with pirates, police, and traders during warp travel
- Banking system with loans and interest
- Crew hiring and management
- Galactic chart with visual star map and sortable/filterable system list
- 10 unique quest lines (Dragonfly, Space Monster, Alien Artifact, and more)
- Wormhole travel network
- Save/load game support
- Runs great on a Raspberry Pi Zero W

## Screenshots

```
+------------------------------------------------------+
|  Beteigeuze                                          |
|  ____________                                        |
|  Tech: Early Industrial  |  Gov: Monarchy            |
|  Credits: 1000  |  Ship: Gnat  |  Hull: 100/100      |
|                                                      |
|  > Market                                            |
|    Short-Range Chart                                 |
|    Shipyard                                          |
|    Bank                                              |
|    Personnel                                         |
|    Galactic Chart                                    |
|    Status                                            |
+------------------------------------------------------+
```

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

| Key       | Action                        |
|-----------|-------------------------------|
| j / k     | Navigate up/down              |
| Enter     | Select / confirm              |
| Esc       | Back / cancel                 |
| 1-5       | Sort by column (charts)       |
| /         | Filter systems (charts)       |
| b / s     | Buy / sell (market)           |
| r         | Refuel (short-range chart)    |
| w         | Use wormhole                  |
| s         | Save game (system hub)        |
| Ctrl+C    | Quit                          |

## Project Structure

```
main.go                  Entry point
tui/
  app.go                 Top-level Bubbletea model and routing
  screens/               One file per game screen
    shared.go            Shared styles, key bindings, helpers
    system_table.go      Sort/filter logic for system lists
    system.go            System hub (main menu per system)
    market.go            Buy/sell goods
    chart.go             Short-range chart (travel)
    galactic_chart.go    Galactic map and system list
    shipyard_screen.go   Ships, equipment, repairs
    bank.go              Loans
    personnel.go         Crew management
    status.go            Commander status
    encounter.go         Combat and encounters
    quest_event.go       Quest narrative events
    newgame.go           Character creation
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
