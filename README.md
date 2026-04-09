```
   _____                          ______               __
  / ___/____  ____ _________     /_  __/________ _____/ /__  _____
  \__ \/ __ \/ __ `/ ___/ _ \     / / / ___/ __ `/ __  / _ \/ ___/
 ___/ / /_/ / /_/ / /__/  __/    / / / /  / /_/ / /_/ /  __/ /
/____/ .___/\__,_/\___/\___/    /_/ /_/   \__,_/\__,_/\___/_/
    /_/
```

A terminal-based port of the classic PalmOS game **Space Trader** by Pieter Spronck, built with Go and [Bubbletea](https://github.com/charmbracelet/bubbletea).

## The Objective

You are a space trader. Buy low, sell high, and travel between star systems to make a profit. Your ultimate goal is to earn 500,000 credits and buy a moon to retire on.

Along the way you will encounter pirates who want your cargo, police who want to inspect it, and a galaxy full of quests, disasters, and opportunities. Some quests are timed. Some are dangerous. Some will change the galaxy permanently if you fail.

There is no single path to victory. You can play as a peaceful merchant, a pirate hunter, a smuggler, or some combination. Your skill choices at the start shape your options.

## How to Play

### Getting started

1. Build and run the game (see below)
2. Create a character: pick a name, difficulty, and allocate skill points
3. You start at a random system with a small ship, a basic weapon, and some credits

### The game loop

From your current system, you can:

- **Trade** -- buy and sell goods at the market. Prices vary by system tech level, government, and random events. The market shows your cost basis so you know what is profitable.
- **Travel** -- open Navigation to see reachable systems. Select one and confirm to warp. You consume fuel equal to the distance. During warp you may encounter pirates, police, or traders.
- **Upgrade** -- visit the Shipyard to buy better ships, weapons, shields, gadgets, repairs, and fuel.
- **Take on quests** -- special events appear as you arrive at systems. Some offer rewards, some are urgent. Check your Status screen to see active quests and their deadlines.
- **Plan routes** -- use the Route Planner (press `p` from Navigation) to find multi-hop paths to distant systems, with refuel costs and trade opportunities along the way.

### Key controls

Every screen shows its available keys at the bottom. The basics:

- `j/k` or arrow keys to navigate menus
- `Enter` to select or confirm
- `Esc` to go back
- `b` to buy (market) or bookmark (navigation)
- `s` to sell (market) or save (system hub)
- `/` to search or filter

### Tips for new players

- Check the **Trader's Guide** from the system menu -- it explains how tech levels, governments, and resource specialties affect prices.
- **Fuel is limited.** If you can't reach a system, you need to refuel at the Shipyard first.
- **Save often.** Press `s` from the system hub.
- When a quest says it has a deadline, it means it. Plan your route before accepting.
- The galaxy map (`L` from Navigation switches between list and map views) helps you see the big picture.

## Build and Run

Requires Go 1.21+ and a terminal with 80+ columns.

```sh
make run
```

Or directly:

```sh
go run .
```

Cross-compile for Raspberry Pi Zero W:

```sh
make build-pi
```

## Credits

Based on [Space Trader](https://www.spronck.net/spacetrader/) by Pieter Spronck, originally released for PalmOS in 2000.

## License

MIT
