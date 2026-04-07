# Space Trader (Palm OS) - Complete Game Mechanics Reference

Original game by Pieter Spronck (2000-2002), GPL licensed.
Data extracted from v1.2.0 source code and the Windows C# port (v2.00).

---

## 1. TRADE GOODS

10 tradeable commodities. Fields: TechProduction (min tech to produce), TechUsage (min tech to use/buy),
TechTopProduction (tech level where production peaks), PriceLowTech (base price at lowest tech),
PriceInc (price change per tech level), Variance (random price fluctuation range),
DoublePriceStatus (system status that increases price 50%), CheapResource (resource that lowers price 25%),
ExpensiveResource (resource that raises price 33%), MinTradePrice, MaxTradePrice, RoundOff.

| Good      | TechProd | TechUse | TechTop | BasePrice | PriceInc | Variance | PriceEvent   | CheapResource  | ExpensiveResource | MinPrice | MaxPrice | Round |
|-----------|----------|---------|---------|-----------|----------|----------|--------------|----------------|-------------------|----------|----------|-------|
| Water     | 0        | 0       | 2       | 30        | +3       | 4        | Drought      | SweetOceans    | Desert            | 30       | 50       | 1     |
| Furs      | 0        | 0       | 0       | 250       | +10      | 10       | Cold         | RichFauna      | Lifeless          | 230      | 280      | 5     |
| Food      | 1        | 0       | 1       | 100       | +5       | 5        | CropFailure  | RichSoil       | PoorSoil          | 90       | 160      | 5     |
| Ore       | 2        | 2       | 3       | 350       | +20      | 10       | War          | MineralRich    | MineralPoor       | 350      | 420      | 10    |
| Games     | 3        | 1       | 6       | 250       | -10      | 5        | Boredom      | Artistic       | (none)            | 160      | 270      | 5     |
| Firearms  | 3        | 1       | 5       | 1250      | -75      | 100      | War          | Warlike        | (none)            | 600      | 1100     | 25    |
| Medicine  | 4        | 1       | 6       | 650       | -20      | 10       | Plague       | SpecialHerbs   | (none)            | 400      | 700      | 25    |
| Machines  | 4        | 3       | 5       | 900       | -30      | 5        | Employment   | (none)         | (none)            | 600      | 800      | 25    |
| Narcotics | 5        | 0       | 5       | 3500      | -125     | 150      | Boredom      | WeirdMushrooms | (none)            | 2000     | 3000     | 50    |
| Robots    | 6        | 4       | 7       | 5000      | -150     | 100      | Employment   | (none)         | (none)            | 3500     | 5000     | 100   |

Firearms and Narcotics are illegal. Systems with politics that ban drugs/firearms will not trade them.

Tech levels: 0=PreAgricultural, 1=Agricultural, 2=Medieval, 3=Renaissance, 4=EarlyIndustrial, 5=Industrial, 6=PostIndustrial, 7=HiTech.

### Price Calculation Formula (StandardPrice)

```
Price = PriceLowTech + (SystemTechLevel * PriceInc)

if Politics.Wanted == this good:
    Price = Price * 4 / 3

Price = Price * (100 - 2 * Politics.StrengthTraders) / 100
Price = Price * (100 - SystemSize) / 100

if SystemResource == CheapResource:
    Price = Price * 3 / 4
if SystemResource == ExpensiveResource:
    Price = Price * 4 / 3

if SystemTechLevel < TechUsage:
    Price = 0  (not sold here)
```

### DeterminePrices (per system, called on arrival)

```
BuyPrice[i] = StandardPrice(...)

if system status matches DoublePriceStatus:
    BuyPrice[i] = BuyPrice[i] * 3 / 2

BuyPrice[i] += Random(Variance) - Random(Variance)

SellPrice[i] = BuyPrice[i]
if PoliceRecordScore < DUBIOUSSCORE:
    SellPrice[i] = SellPrice[i] * 90 / 100   (criminals get 10% less)

RecalculateBuyPrices()  (applies trader skill discount)
```

### Trader Skill Effect on Buy Prices

```
BuyPrice = BuyPrice * (103 + (MAXSKILL - TraderSkill)) / 100
```

With max trader skill (10): pay 103% of base. With min (1): pay 112%.

### Sell Price

Base sell price is 75% of standard price:
```
BaseSellPrice = Price * 3 / 4
```

### Trade Item Quantity Initialization

```
Qty = (Random(9..13) - |TechTopProduction - SystemTechLevel|) * (1 + SystemSize)

if Narcotics or Robots:
    Qty = (Qty * (5 - Difficulty)) / (6 - Difficulty) + 1

if CheapResource matches: Qty = Qty * 4 / 3
if ExpensiveResource matches: Qty = Qty * 3 / 4
if DoublePriceStatus active: Qty = Qty / 5

Qty += Random(10) - Random(10)
if Qty < 0: Qty = 0
```

### Dumping and Jettisoning Cargo

- Dump cargo cost: 5 * (Difficulty + 1) credits per unit
- Jettison cargo: free but illegal, reduces police record

---

## 2. SHIP TYPES

Fields: Name, CargoBays, WeaponSlots, ShieldSlots, GadgetSlots, CrewQuarters, FuelTanks,
MinTechLevel, FuelCost, BasePrice, Bounty, Occurrence%, HullStrength,
PoliceMin, PirateMin, TraderMin, Size, RepairCost.

| Ship        | Cargo | Wpn | Shd | Gad | Crew | Fuel | Tech | FuelCost | Price   | Bounty | Occur% | Hull | Police | Pirate | Trader | Size | Repair |
|-------------|-------|-----|-----|------|------|------|------|----------|---------|--------|--------|------|--------|--------|--------|------|--------|
| Flea        | 10    | 0   | 0   | 0    | 1    | 20   | 4    | 1        | 2000    | 5      | 2%     | 25   | -1     | -1     | 0      | 0    | 1      |
| Gnat        | 15    | 1   | 0   | 1    | 1    | 14   | 5    | 2        | 10000   | 50     | 28%    | 100  | 0      | 0      | 0      | 1    | 1      |
| Firefly     | 20    | 1   | 1   | 1    | 1    | 17   | 5    | 3        | 25000   | 75     | 20%    | 100  | 0      | 0      | 0      | 1    | 1      |
| Mosquito    | 15    | 2   | 1   | 1    | 1    | 13   | 5    | 5        | 30000   | 100    | 20%    | 100  | 0      | 1      | 0      | 1    | 1      |
| Bumblebee   | 25    | 1   | 2   | 2    | 2    | 15   | 5    | 7        | 60000   | 125    | 15%    | 100  | 1      | 1      | 0      | 1    | 2      |
| Beetle      | 50    | 0   | 1   | 1    | 3    | 14   | 5    | 10       | 80000   | 50     | 3%     | 50   | -1     | -1     | 0      | 1    | 2      |
| Hornet      | 20    | 3   | 2   | 1    | 2    | 16   | 6    | 15       | 100000  | 200    | 6%     | 150  | 2      | 3      | 1      | 2    | 3      |
| Grasshopper | 30    | 2   | 2   | 3    | 3    | 15   | 6    | 15       | 150000  | 300    | 2%     | 150  | 3      | 4      | 2      | 3    | 3      |
| Termite     | 60    | 1   | 3   | 2    | 3    | 13   | 7    | 20       | 225000  | 300    | 2%     | 200  | 4      | 5      | 3      | 4    | 4      |
| Wasp        | 35    | 3   | 2   | 2    | 3    | 14   | 7    | 20       | 300000  | 500    | 2%     | 200  | 5      | 6      | 4      | 5    | 4      |

Special (non-purchasable) ships:

| Ship          | Cargo | Wpn | Shd | Gad | Crew | Fuel | Hull | Notes              |
|---------------|-------|-----|-----|------|------|------|------|--------------------|
| Space Monster | 0     | 3   | 0   | 0    | 1    | 1    | 500  | Quest enemy        |
| Dragonfly     | 0     | 2   | 3   | 2    | 1    | 1    | 10   | Quest enemy        |
| Mantis        | 0     | 3   | 1   | 3    | 3    | 1    | 300  | Random encounter   |
| Scarab        | 20    | 2   | 0   | 0    | 2    | 1    | 400  | Quest enemy        |
| Bottle        | 0     | 0   | 0   | 0    | 0    | 1    | 10   | Message in a bottle|

Size values: 0=Tiny, 1=Small, 2=Medium, 3=Large, 4=Huge.

Flea has MAXRANGE (20) fuel -- highest range of any ship.
Encounters are half as likely in a Flea.

### Ship Trade-In Value

```
TradeIn = BasePrice * 3/4   (or 1/4 if tribbles infesting)
TradeIn -= RepairCost * (MaxHull - CurrentHull)
TradeIn -= FuelCost * (MaxFuel - CurrentFuel)
TradeIn += 2/3 * (sum of weapon prices)
TradeIn += 2/3 * (sum of shield prices)
TradeIn += 2/3 * (sum of gadget prices)
```

### Buying a New Ship

```
ShipPrice = BaseShipPrice - CurrentShipTradeInValue
```
- Must be at system with MinTechLevel >= ship's required tech
- Cannot buy if in debt
- Special equipment (Lightning Shield, Fuel Compactor, Morgan's Laser) transfers if new ship has slots
- Transfer costs added to price

---

## 3. EQUIPMENT

### Weapons

| Weapon         | Power | Price  | MinTech       | Chance% |
|----------------|-------|--------|---------------|---------|
| Pulse Laser    | 15    | 2000   | Industrial(5) | 50      |
| Beam Laser     | 25    | 12500  | PostInd(6)    | 35      |
| Military Laser | 35    | 35000  | HiTech(7)     | 15      |
| Morgan's Laser | 85    | 50000  | Unavailable   | 0       |

Morgan's Laser is a quest reward only (reactor delivery to Nix).

Equipment sell price = 2/3 of purchase price (for all equipment types).

### Shields

| Shield           | Power | Price  | MinTech       | Chance% |
|------------------|-------|--------|---------------|---------|
| Energy Shield    | 100   | 5000   | Industrial(5) | 70      |
| Reflective Shield| 200   | 20000  | PostInd(6)    | 30      |
| Lightning Shield | 350   | 45000  | Unavailable   | 0       |

Lightning Shield is a quest reward (Dragonfly destruction at Zalkon).

### Gadgets

| Gadget           | Price   | MinTech         | Chance% | Effect                          |
|------------------|---------|-----------------|---------|----------------------------------|
| 5 Extra Cargo Bays| 2500   | EarlyInd(4)     | 35      | +5 cargo capacity (stackable)   |
| Auto-Repair System| 7500   | Industrial(5)   | 20      | +3 engineer skill bonus         |
| Navigating System| 15000   | PostInd(6)      | 20      | +3 pilot skill, +2 cloak bonus  |
| Targeting System | 25000   | PostInd(6)      | 20      | +3 fighter skill bonus          |
| Cloaking Device  | 100000  | HiTech(7)       | 5       | +2 pilot skill bonus (cloaking) |
| Fuel Compactor   | 30000   | Unavailable     | 0       | Sets fuel capacity to 18        |

Fuel Compactor is a quest reward (Gemulon rescue).
Each gadget type can only be installed once per ship, EXCEPT Extra Cargo Bays.

SkillBonus = 3, CloakBonus = 2.

### Escape Pod

- Cost: 2000 credits
- Ejects crew when hull reaches 0
- Required for insurance
- After ejection: player gets a Flea, loses cargo and equipment
- Transfers between ships

---

## 4. PLAYER SKILLS

4 skills: Pilot, Fighter, Trader, Engineer. Max value: 10.
Starting skill points: 16 total to distribute across 4 skills.
Random skill generation for mercenaries: 1 + Random(0..4) + Random(0..5), range 1-10.

### Skill Calculation (effective skill = best crew member + gadget bonuses)

```
PilotSkill   = max(crew pilot skills) + NavigatingSystem(+3) + CloakingDevice(+2)
FighterSkill = max(crew fighter skills) + TargetingSystem(+3)
TraderSkill  = max(crew trader skills)   (no gadget bonus)
EngineerSkill= max(crew engineer skills) + AutoRepairSystem(+3)
```

### Difficulty Adjustment

```
Beginner/Easy: skill + 1
Normal: no change
Hard: no change (code shows no adjustment)
Impossible: max(1, skill - 1)
```

### What Each Skill Does

- Pilot: Flee success probability, dodge attacks, pursue fleeing enemies
- Fighter: Hit probability in combat, weapon accuracy
- Trader: Reduces buy prices (formula: BuyPrice * (103 + (10 - TraderSkill)) / 100)
- Engineer: Repairs hull/shields during travel, reduces damage taken, boosts weapon damage

### Skill Increase Events

- Quest reward (Medicine delivery to Japori): 2 random skill increases
- Skill Increase special event: 1 random skill increase (costs 3000 credits)
- Bottle (Good): random skill tweaks based on difficulty

---

## 5. CREW MEMBERS

### Mercenary Hire Price (daily)

```
HirePrice = (Pilot + Fighter + Trader + Engineer) * 3
```

Paid before each warp. If player cannot afford, mercenary leaves.

### Named Crew Members

Commander (player), plus 30 hirable mercenaries:
Alyssa, Armatur, Bentos, C2U2, Chi'Ti, Crystal, Dane, Deirdre, Doc, Draco,
Iranda, Jeremiah, Jujubal, Krydon, Luis, Mercedez, Milete, Muri-L, Mystyc,
Nandi, Orestes, Pancho, PS37, Quarck, Sosumi, Uma, Wesley, Wonton, Yorvick, Zeethibal.

Regular mercenary skills are randomly generated and assigned to random systems.
One mercenary available per system (found via Personnel Roster).

### Special Crew Members (quest-related)

| Name           | Pilot | Fighter | Trader | Engineer | Notes                    |
|----------------|-------|---------|--------|----------|--------------------------|
| Zeethibal      | 5     | 5       | 5      | 5        | Rate: 0 (free)           |
| Wild           | 7     | 10      | 2      | 5        | Smuggle to Kravat        |
| Jarek          | 3     | 2       | 10     | 4        | Transport to Devidia     |
| FamousCaptain  | 10    | 10      | 10     | 10       | Temporary encounter      |

### Quest Enemy Crew Skills (scale with difficulty d=0..4)

| Enemy       | Pilot | Fighter | Trader | Engineer |
|-------------|-------|---------|--------|----------|
| Dragonfly   | 4+d   | 6+d     | 1      | 6+d      |
| Scarab      | 5+d   | 6+d     | 1      | 6+d      |
| SpaceMonster| 8+d   | 8+d     | 1      | 1+d      |

---

## 6. ENCOUNTERS

### Encounter Generation

Each "click" during travel (21 clicks total per warp), the game rolls:
```
EncounterTest = Random(44 - 2 * Difficulty)
if ship is Flea: EncounterTest *= 2  (half as likely)
```

Then checks thresholds in order:
1. If EncounterTest < PirateStrength: pirate encounter
2. If EncounterTest < PirateStrength + PoliceStrength: police encounter
3. If EncounterTest < PirateStrength + PoliceStrength + TraderStrength: trader encounter

Pirate/Police/Trader strength comes from the destination system's political system.

### Special Encounter Chances (per click)

- Marie Celeste: 1/1000 chance (derelict ship with narcotics)
- Captain Ahab: requires reflective shield, Pilot < 10, record > Criminal
- Captain Conrad: requires military laser, Engineer < 10
- Captain Huie: requires military laser, Trader < 10
- Bottle (skill tonic): 1/1000 chance
- Mantis: if carrying alien artifact, 4/20 chance; at Gemulon during invasion, >4/10 chance

### Specific Quest Encounters

| Location | Clicks | Condition          | Enemy         |
|----------|--------|--------------------|---------------|
| Acamar   | 1      | MonsterStatus == 1 | Space Monster |
| Zalkon   | 1      | DragonflyStatus==4 | Dragonfly     |
| Wormhole | 20     | ScarabStatus == 1  | Scarab        |

### Combat Mechanics

#### Hit Probability

```
HitChance = FighterSkill(Attacker) + ShipSize(Defender)
vs
DodgeChance = 5 + (PilotSkill(Defender) / 2)

if defender is fleeing: DodgeChance /= 2
```

#### Damage Formula

```
Damage = Random(TotalWeaponPower * (100 + 2 * EngineerSkill) / 100)
```

Against Scarab: only Pulse Laser and Morgan's Laser deal damage.

#### Shield Absorption

Shields absorb damage sequentially. Only after all shields depleted does hull take damage.

#### Hull Damage Reduction

```
HullDamage -= Random(EngineerSkill(Defender))

Minimum shots to destroy (hull damage cap per hit):
  Beginner:   hull / 4
  Easy:       hull / 3
  Normal:     hull / 2
  Hard/Impossible: hull (no cap, can one-shot)
```

#### Flee Mechanics

Commander fleeing:
```
Success = (Random(7) + PilotSkill/3) * 2 >= Random(OpponentPilot) * (2 + Difficulty)
```
On Beginner: always escape unharmed.

Opponent fleeing:
```
Success = Random(PlayerPilot) * 4 <= Random(7 + OpponentPilot/3) * 2
```

#### Surrender Mechanics

Arrest fine: ((1 + (Worth * min(80, -PoliceRecord) / 100) / 500)) * 500
Prison time: max(30, -PoliceRecordScore) days
Police record reset to DUBIOUS (-5) after arrest.
If Wild aboard: fine * 1.05.

#### Bribe Formula

```
BribeCost = Worth / ((10 + 5 * (4 - Difficulty)) * Politics.BribeLevel)
Rounded to nearest 100, min 100, max 10000.
Doubled if Wild or Reactor aboard.
```

Cannot bribe in systems where BribeLevel <= 0 (Fascist, Military, Technocracy, Theocracy).

#### Bounty Calculation

```
Bounty = EnemyShipPrice / 200
Rounded down to nearest 25
Min: 25, Max: 2500
```

Where EnemyShipPrice = BasePrice + WeaponPrices + ShieldPrices, scaled by enemy skills.

#### Scoop/Loot Probability

After destroying a ship, chance to scoop cargo:
- Beginner/Easy: 100%
- Normal: 50%
- Hard: 33%
- Impossible: 25%

---

## 7. POLICE RECORD

| Score  | Value | Label      |
|--------|-------|------------|
| < -100 | -100  | Psychopath |
| < -70  | -70   | Villain    |
| < -30  | -30   | Criminal   |
| < -10  | -10   | Crook      |
| < -5   | -5    | Dubious    |
| 0      | 0     | Clean      |
| > 5    | 5     | Lawful     |
| > 10   | 10    | Trusted    |
| > 25   | 25    | Liked      |
| > 75   | 75    | Hero       |

### Score Changes

| Action          | Change |
|-----------------|--------|
| Attack Police   | -3     |
| Kill Police     | -6     |
| Kill Trader     | -4     |
| Kill Pirate     | +1     |
| Plunder Trader  | varies |
| Trafficking     | varies |

### Police Record Decay (per warp day)

```
if score > CLEAN(0): decrement every 3 days (gravitates toward neutral)
if score < DUBIOUS(-5) and Difficulty <= Normal: increment by 1 every day
if score < DUBIOUS(-5) and Difficulty > Normal: increment by 1 every (Difficulty) days
```

---

## 8. REPUTATION (COMBAT RATING)

| Kills  | Value | Label           |
|--------|-------|-----------------|
| 0      | 0     | Harmless        |
| 10     | 10    | Mostly Harmless |
| 20     | 20    | Poor            |
| 40     | 40    | Average         |
| 80     | 80    | Above Average   |
| 150    | 150   | Competent       |
| 300    | 300   | Dangerous       |
| 600    | 600   | Deadly          |
| 1500   | 1500  | Elite           |

ReputationScore incremented by kills (pirate, police, trader combined).

---

## 9. POLITICAL SYSTEMS

Fields: Name, MinTechLevel, MaxTechLevel, StrengthPolice, StrengthPirates,
StrengthTraders, BribeLevel, MinTechLevel(unused?), DrugsOK, FirearmsOK, WantedTradeItem.

| Government       | MinTech | MaxTech | Police | Pirates | Traders | Bribe | Drugs | Firearms | Wanted   |
|------------------|---------|---------|--------|---------|---------|-------|-------|----------|----------|
| Anarchy          | 0       | 0       | 7      | 1       | 0       | 5     | Yes   | Yes      | Food     |
| Capitalist State | 2       | 3       | 2      | 7       | 4       | 7     | Yes   | Yes      | Ore      |
| Communist State  | 6       | 6       | 4      | 4       | 1       | 5     | Yes   | Yes      | (none)   |
| Confederacy      | 5       | 4       | 3      | 5       | 1       | 6     | Yes   | Yes      | Games    |
| Corporate State  | 2       | 6       | 2      | 7       | 4       | 7     | Yes   | Yes      | Robots   |
| Cybernetic State | 0       | 7       | 7      | 5       | 6       | 7     | No    | No       | Ore      |
| Democracy        | 4       | 3       | 2      | 5       | 3       | 7     | Yes   | Yes      | Games    |
| Dictatorship     | 3       | 4       | 5      | 3       | 0       | 7     | Yes   | Yes      | (none)   |
| Fascist State    | 7       | 7       | 7      | 1       | 4       | 7     | No    | Yes      | Machines |
| Feudal State     | 1       | 1       | 6      | 2       | 0       | 3     | Yes   | Yes      | Firearms |
| Military State   | 7       | 7       | 0      | 6       | 2       | 7     | No    | Yes      | Robots   |
| Monarchy         | 3       | 4       | 3      | 4       | 0       | 5     | Yes   | Yes      | Medicine |
| Pacifist State   | 7       | 2       | 1      | 5       | 0       | 3     | Yes   | No       | (none)   |
| Socialist State  | 4       | 2       | 5      | 3       | 0       | 5     | Yes   | Yes      | (none)   |
| State of Satori  | 0       | 1       | 1      | 1       | 0       | 1     | No    | No       | (none)   |
| Technocracy      | 1       | 6       | 3      | 6       | 4       | 7     | Yes   | Yes      | Water    |
| Theocracy        | 5       | 6       | 1      | 4       | 0       | 4     | Yes   | Yes      | Narcotics|

Note: Police and Pirates columns are activity levels (0-7 scale), where higher = more active.
BribeLevel: higher = harder to bribe. 0 = impossible.
Wanted: the trade item this government pays premium for (price * 4/3).

---

## 10. SPECIAL RESOURCES (System Attributes)

| ID | Resource        | Effect                                    |
|----|-----------------|-------------------------------------------|
| 0  | Nothing Special | No effect                                 |
| 1  | Mineral Rich    | Ore cheaper (CheapResource)               |
| 2  | Mineral Poor    | Ore more expensive (ExpensiveResource)     |
| 3  | Desert          | Water more expensive                       |
| 4  | Sweetwater Oceans| Water cheaper                             |
| 5  | Rich Soil       | Food cheaper                               |
| 6  | Poor Soil       | Food more expensive                        |
| 7  | Rich Fauna      | Furs cheaper                               |
| 8  | Lifeless        | Furs more expensive                        |
| 9  | Weird Mushrooms | Narcotics cheaper                          |
| 10 | Special Herbs   | Medicine cheaper                           |
| 11 | Artistic        | Games cheaper                              |
| 12 | Warlike         | Firearms cheaper                           |

40% chance a system has a special resource (60% have nothing special).

---

## 11. SYSTEM STATUS/PRESSURE EVENTS

| ID | Status          | Effect on prices                           |
|----|-----------------|---------------------------------------------|
| 0  | Uneventful      | No effect                                   |
| 1  | War             | Ore and Firearms prices * 1.5, quantities / 5 |
| 2  | Plague          | Medicine prices * 1.5, quantities / 5       |
| 3  | Drought         | Water prices * 1.5, quantities / 5          |
| 4  | Boredom         | Games and Narcotics prices * 1.5            |
| 5  | Cold            | Furs prices * 1.5, quantities / 5           |
| 6  | Crop Failure    | Food prices * 1.5, quantities / 5           |
| 7  | Lack of Workers | Machines and Robots prices * 1.5            |

15% chance a system has a status event at game start.
Status shuffles with 15% probability per system per warp.

---

## 12. SYSTEM SIZES

| Value | Name   |
|-------|--------|
| 0     | Tiny   |
| 1     | Small  |
| 2     | Medium |
| 3     | Large  |
| 4     | Huge   |

System size affects: price (larger = slightly cheaper), trade quantities (larger = more goods).

---

## 13. GALAXY GENERATION

- Galaxy dimensions: 150 x 110 (some sources say 154 x 110 for Windows version)
- Number of solar systems: 120 (Palm), up to 120
- Number of wormholes: 6 (MAXWORMHOLE), forming a circular chain
- Minimum distance between systems: 6 parsecs
- Close distance threshold: 13 parsecs (each system must have at least one neighbor within this)
- Wormhole distance: 3 parsecs

### System Generation Algorithm

1. First MAXWORMHOLE systems are placed with wormhole-friendly positions
2. Remaining systems placed randomly with minimum distance constraints
3. Each system must have at least one neighbor within CLOSEDISTANCE
4. Tech level: random 0-7
5. Politics: random, but must be compatible with tech level (min/max tech constraints)
6. Special resources: 40% chance of having one
7. Size: random 0-4
8. Status: 15% chance of non-uneventful status

### Named Systems (120+)

Acamar, Adahn, Aldea, Andevian, Antedi, Balosnee, Baratas, Brax, Bretel, Calondia,
Campor, Capelle, Carzon, Castor, Cestus, Cheron, Courteney, Daled, Damast, Davlos,
Deneb, Deneva, Devidia, Draylon, Drema, Endor, Esmee, Exo, Ferris, Festen, Fourmi,
Frolix, Gemulon, Guinifer, Hades, Hamlet, Helena, Hulst, Iodine, Iralius, Janus,
Japori, Jarada, Jason, Kaylon, Khefka, Kira, Klaatu, Klaestron, Korma, Kravat, Krios,
Laertes, Largo, Lave, Ligon, Lowry, Magrat, Malcoria, Melina, Mentar, Merik, Mintaka,
Montor, Mordan, Myrthe, Nelvana, Nix, Nyle, Odet, Og, Omega, Omphalos, Orias, Othello,
Parade, Penthara, Picard, Pollux, Quator, Rakhar, Ran, Regulas, Relva, Rhymus, Rochani,
Rubicum, Rutia, Sarpeidon, Sefalla, Seltrice, Sigma, Sol, Somari, Stakoron, Styris,
Talani, Tamus, Tantalos, Tanuga, Tarchannen, Terosa, Thera, Titan, Torin, Triacus,
Turkana, Tyrus, Umberlee, Utopia, Vadera, Vagra, Vandor, Ventax, Xenon, Xerxes, Yew,
Yojimbo, Zalkon, Zuul.

Quest-critical systems: Acamar (monster), Baratas/Melina/Regulas/Zalkon (dragonfly),
Japori (disease), Kravat (Wild), Devidia (Jarek), Gemulon (invasion),
Daled (experiment), Nix (reactor/laser), Utopia (moon).

---

## 14. WARP/TRAVEL MECHANICS

### Fuel

- Fuel measured in parsecs of range
- Cost per parsec varies by ship type (FuelCost field)
- Wormhole travel: no fuel consumed, but costs WormholeTax
- WormholeTax = ShipType.FuelCost * 25

### Travel Sequence (21 clicks)

Each warp takes 1 day. During travel (21 clicks counting down):

1. Engineer repairs hull: Random(EngineerSkill) / 2 per click
2. Surplus repair goes to shields at 2x rate
3. Encounter checks at each click
4. Shields fully recharged before departure

### Per-Warp Deductions

- Fuel consumed (distance in parsecs)
- Mercenary wages
- Insurance premium
- Interest on debt
- Wormhole tax (if applicable)

### Insurance Cost Formula

```
InsuranceMoney = max(1, (ShipPrice * 5 / 2000) * (100 - min(NoClaim, 90)) / 100)
```

Insurance rate: 0.25% of ship value per day, reduced by no-claim bonus (1% per day, max 90%).

### Interest on Debt

```
Interest = max(1, Debt / 10)   (10% of outstanding debt per warp)
if Credits >= Interest: Credits -= Interest
else: Debt += (Interest - Credits); Credits = 0
```

### Police Record Decay During Travel

```
if record > CLEAN: decrease by 1 every 3 days
if record < DUBIOUS:
  Normal or easier: increase by 1 per day
  Hard: increase by 1 every 4 days
  Impossible: increase by 1 every 5 days
```

### Space Monster Regeneration

```
MonsterHull = MonsterHull * 105 / 100  (5% per day, capped at max)
```

---

## 15. SPECIAL QUESTS

### Space Monster (Acamar)

- Accept quest at a random system
- Travel to Acamar to fight the Space Monster (500 hull, 3 weapons, very tough)
- Monster regenerates 5% hull per day
- Reward: -15000 (you PAY 15000? or receive -- source shows negative = reward)

### Dragonfly Chase (Baratas -> Melina -> Regulas -> Zalkon)

- Follow the Dragonfly through 4 systems
- Final battle at Zalkon (Dragonfly has only 10 hull but 3 shields)
- Reward: Lightning Shield (350 power, worth 45000)

### Japori Disease (Japori)

- Accept quest, receive 10 bays of medicine (occupies cargo space)
- Deliver to Japori
- Reward: 2 random skill increases

### Alien Artifact

- Find artifact at a random system, deliver to professor Berger at a Hi-Tech system
- While carrying artifact, Mantis encounters increase significantly (4/20 chance per click)
- Reward: 20000 credits

### Ambassador Jarek (Devidia)

- Transport Jarek to Devidia
- Jarek acts as crew member (Pilot 3, Fighter 2, Trader 10, Engineer 4)
- No cash reward, but Jarek's trader skill is very helpful during transport

### Jonathan Wild (Kravat)

- Smuggle criminal Wild to Kravat
- Wild acts as crew (Pilot 7, Fighter 10, Trader 2, Engineer 5)
- Police encounters more dangerous while Wild aboard
- Requires beam laser to keep Wild under control
- Reward: leads to Morgan's Reactor quest

### Morgan's Reactor (Nix)

- Receive unstable reactor after Wild delivery
- Must deliver to Nix before it melts down (status 1-20, melts at 21)
- Reactor occupies 5 bays + diminishing enriched fuel (10 - (status-1)/2 bays)
- Reactor halves tribble population per warp
- Reward: Morgan's Laser (85 power)

### Alien Invasion (Gemulon)

- Warning about impending invasion
- Must reach Gemulon within ~7 days
- If failed: Gemulon becomes TechLevel 0, Anarchy
- Reward: Fuel Compactor gadget (sets fuel to 18)

### Dangerous Experiment (Daled)

- Stop Dr. Fehler's experiment at Daled within ~11 days
- If failed: "fabric rip" causes random warps for some time
- FabricRipProbability starts at 25, decreases by 1 per day

### Scarab (wormhole exit)

- Find and destroy the Scarab at a wormhole exit
- Scarab has 400 hull, can ONLY be damaged by Pulse Laser and Morgan's Laser
- Reward: Hull upgrade (hardened hull, +50% hull strength)

### Moon Purchase (Utopia)

- Available when net worth >= 500000 credits
- Moon costs 500000 credits
- Claim moon at Utopia to end the game
- COSTMOON = 500000

### Skill Increase

- Available at random systems
- Cost: 3000 credits
- Reward: 1 random skill increase

### Erase Record

- Available at random systems
- Cost: 5000 credits
- Resets police record to Clean (0)

### Tribble Buyer

- Available at 3 random systems
- Buys all your tribbles for credits

### Cargo For Sale

- Random event at systems
- Cost: 1000 credits
- Receive random cheap cargo

### Lottery Winner

- Random event
- Receive 1000 credits

### Merchant Prince

- Random event
- Cost: 1000 credits
- Reward: trader skill related

---

## 16. TRIBBLES

- First acquired through random encounter or purchase
- Breed every warp: Tribbles += 1 + Random(max(1, Tribbles / (FoodOnBoard ? 1 : 2)))
- If food aboard: Tribbles += 100 + Random(Cargo[Food] * 100), food consumed
- If narcotics aboard: Tribbles = 1 + Random(3), narcotics consumed, converted to Furs
- If reactor aboard: Tribbles /= 2 per warp
- Max tribbles: 100,000
- Tribble infestation reduces ship trade-in to 1/4 value
- Can sell to Tribble Buyer at designated systems
- Visual display during encounters: sqrt(Tribbles / 250) tribbles shown

---

## 17. BANKING

### Loans

```
MaxLoan (clean record): min(25000, max(1000, (Worth / 10 / 500) * 500))
MaxLoan (criminal):     500
```

### Interest

- 10% of outstanding debt per warp (deducted from credits first, else added to debt)
- DebtWarning at 75000 credits debt
- DebtTooLarge at 100000 credits debt (cannot warp)

### Debt Effects

- Reminder every 5 days if debt > 0
- Cannot buy ships if in debt
- Cannot warp if debt >= 100000

---

## 18. INSURANCE

- Requires escape pod
- Daily cost: max(1, (ShipPrice * 5 / 2000) * (100 - min(NoClaim, 90)) / 100)
- No-claim bonus: +1% per day, max 90%
- Transfers to new ships
- Resets no-claim to 0 after pod ejection
- Pays out ship value (without cargo) if ship destroyed with pod

---

## 19. DIFFICULTY LEVELS

| Level      | Value | Effects                                              |
|------------|-------|------------------------------------------------------|
| Beginner   | 0     | +1 all skills, guaranteed flee, 4 shots to kill hull |
| Easy       | 1     | +1 all skills, 3 shots to kill hull, 100% scoop     |
| Normal     | 2     | Standard, 2 shots to kill hull, 50% scoop            |
| Hard       | 3     | -0 skills, 1 shot can kill hull, 33% scoop           |
| Impossible | 4     | -1 all skills, 1 shot can kill hull, 25% scoop       |

Additional difficulty effects:
- Encounter frequency: Random(44 - 2*Difficulty) -- higher difficulty = more encounters
- Police record decay slower on Hard/Impossible
- Quest enemy skills: base + difficulty value
- Narcotics/Robots quantities reduced on higher difficulties
- Bribe costs increase with difficulty
- Flee probability harder at higher difficulty

---

## 20. SCORING

Score is calculated on game end (moon purchase, death, or retirement).
Baseline: 100% = claiming a moon in average time with average money on Normal difficulty.
Hard/Impossible can exceed 300%.

Game end types:
- Killed (ship destroyed, no escape pod)
- Retired (quit the game)
- Moon (purchased moon at Utopia -- best ending)

The exact formula references: Score = f(Days, Worth, Difficulty, GameEndType)
Details from source: score stored as integer (displayed / 10 with one decimal).

---

## 21. NEWS SYSTEM

Each system has a newspaper with mastheads based on political system:

| Government       | Mastheads                                           |
|------------------|-----------------------------------------------------|
| Anarchy          | The Arsenal, The Grassroot, Kick It!                |
| Capitalist State | The Objectivist, The Market, The Invisible Hand     |
| Communist State  | The Daily Worker, The People's Voice, The Proletariat|
| Confederacy      | Planet News, The Times, Interstate Update           |
| Corporate State  | Memo, News From The Board, Status Report            |
| Cybernetic State | Pulses, Binary Stream, The System Clock             |
| Democracy        | The Daily Planet, The Majority, Unanimity           |
| Dictatorship     | The Command, Leader's Voice, The Mandate            |
| Fascist State    | State Tribune, Motherland News, Homeland Report     |
| Feudal State     | News from the Keep, The Town Crier, The Herald      |
| Military State   | General Report, Dispatch, The Sentry                |
| Monarchy         | Royal Times, The Loyal Subject, The Fanfare         |
| Pacifist State   | Pax Humani, Principle, The Chorus                   |
| Socialist State  | All for One, Brotherhood, The People's Syndicate    |
| State of Satori  | The Daily Koan, Haiku, One Hand Clapping            |
| Technocracy      | The Future, Hardware Dispatch, TechNews             |
| Theocracy        | The Spiritual Advisor, Church Tidings, Temple Tribune|

News reports include:
- Quest progress updates (dragonfly sightings, monster kills, etc.)
- Player reputation news (villain warnings, hero celebrations)
- System status events (war, plague, drought, etc.)
- Nearby system status events
- Generic filler headlines

News events tracked: ArtifactDelivery, CaughtLittering, Dragonfly sightings, ExperimentFailed/Performed/Stopped, GemulonInvaded/Rescued, Captain encounters (Ahab/Conrad/Huie attacked/destroyed), Japori delivery, JarekGetsOut, Scarab events, SpaceMonster killed, WildArrested/GetsOut.

---

## 22. SPECIAL CARGO

Items that occupy cargo/crew space during quests:

| Cargo              | Bays Used | Condition              |
|--------------------|-----------|------------------------|
| Japori Antidote     | 10        | JaporiDiseaseStatus==1 |
| Alien Artifact      | 1         | ArtifactOnBoard==true  |
| Unstable Reactor    | 5 + fuel  | ReactorStatus 1-20     |
| Reactor Fuel        | 10-(status-1)/2 | Diminishes over time |
| Tribbles            | 0         | Ship.Tribbles > 0      |
| Jonathan Wild       | 1 crew    | WildStatus == 1        |
| Ambassador Jarek    | 1 crew    | JarekStatus == 1       |
| Portable Singularity| 0         | CanSuperWarp flag      |

---

## 23. FAMOUS CAPTAIN ENCOUNTERS

| Captain       | Condition                              | Reward              |
|---------------|----------------------------------------|----------------------|
| Captain Ahab  | Has reflective shield, Pilot < 10     | +1 Pilot skill       |
| Captain Conrad| Has military laser, Engineer < 10      | +1 Engineer skill    |
| Captain Huie  | Has military laser, Trader < 10        | +1 Trader skill      |

Also requires reputation > Criminal score.

---

## 24. MARIE CELESTE

- 1/1000 chance encounter during travel
- Derelict ship containing narcotics cargo
- Player can take the cargo (free narcotics)
- Police may investigate afterward (Marie Celeste police encounter)

---

## 25. CHEAT CODES

Three base cheat codes (Caesar cipher encoded in source).
Version 1.2.0 additions: teleportation, encounter control.

---

## Sources

- Original Palm OS source code: https://github.com/videogamepreservation/spacetrader
- Alternative source mirror: https://github.com/adambair/spacetrader
- Windows C# port: https://github.com/SpaceTraderGame/SpaceTrader-Windows
- C# port by pmprog: https://github.com/pmprog/SpaceTrader
- Java refactored version: https://github.com/LeonisX/space-trader
- Official FAQ mirror: https://memalign.github.io/m/spacetrader-official-site-mirror/STFAQ.html
- GameFAQs strategy guide: https://gamefaqs.gamespot.com/palmos/917550-space-trader/faqs/23321
- Original help file mirror: https://stuff.mit.edu/afs/sipb/user/golem/tmp/pilot/SpaceTrader/SpaceTrader.html
