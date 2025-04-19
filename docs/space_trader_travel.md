# Travel & Encounters Specification

## Overview
Travel is the process of moving between star systems. Encounters occur during travel and may involve traders, pirates, or police.

## Travel Mechanics
- Select destination from short-range chart (limited by fuel/tank)
- Steps:
  1. Leave spaceport (prep: refuel, repair, buy/sell)
  2. Warp (consumes fuel)
  3. Encounter phase (random events)
  4. Approach/dock at destination
- Wormholes: Instant travel, fee based on ship size, no fuel used

## Encounters
- Types: Trader, Pirate, Police
- Trader: May offer trades, can be attacked, may surrender/flee
- Pirate: Usually attacks, may flee if player is strong, can surrender
- Police: May inspect, can be bribed (except incorruptible), attack criminals

## Combat
- Options: Attack, flee, surrender, bribe (contextual)
- Outcomes: Surrender (cargo/cash lost), destruction (escape pod possible), loot, fines, prison
- Combat influenced by skills, ship stats, equipment

## Police & Reputation
- Inspections for illegal goods
- Bribery cost varies by government
- Police record affects encounter frequency and severity
- Severe criminal record: police attack on sight

## Persistent State
- Travel history, encounters, and reputation persist across saves/loads

---
[Back to PRD](space_trader_prd.md)
