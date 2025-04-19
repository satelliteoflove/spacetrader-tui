# Trading System Specification

## Overview
The trading system is the economic core of Space Trader. Players buy and sell goods at spaceports in different star systems, seeking profit by exploiting price differences. The system is influenced by tech levels, political systems, special events, and random fluctuations.

## Trade Goods
- 10 goods: Water, Furs, Food, Ore, Games, Firearms (illegal), Medicine, Machines, Narcotics (illegal), Robots
- Each good has a base price, min/max price, and is affected by:
  - System tech level
  - Political system
  - Special resources/events (e.g., drought, war)
  - Random daily fluctuation

## Buying & Selling
- Goods are bought/sold at the spaceport market.
- Player can buy as much as cargo space and cash allow.
- Selling price is determined by current system state.
- Illegal goods (Firearms, Narcotics) risk police inspection/confiscation.

## Price Calculation
- Base price modified by:
  - Tech level (natural goods cheaper in low-tech, industrial/hi-tech goods cheaper in hi-tech)
  - Political system (some restrict/ban certain goods)
  - Special events (e.g., war increases ore/firearms price)
  - Special resources (e.g., mineral-rich system lowers ore price)
  - Player's Trader skill (up to 10% discount)
  - Random fluctuation (e.g., ±10%)

## Special Events
- Events (war, drought, crop failure, plague, boredom, etc.) can temporarily spike or crash prices.
- Events last several days and are communicated via in-game news.

## Illegal Goods
- Carrying illegal goods risks police inspection.
- On inspection: goods confiscated, player fined, police record worsens.
- Bribery possible depending on government type.

## Loans
- Player may take loans (10% daily interest, auto-deducted on warp).
- Defaulting leads to ship repossession.

## Persistent State
- Market prices, events, and inventory persist across saves/loads.

---
[Back to PRD](space_trader_prd.md)
