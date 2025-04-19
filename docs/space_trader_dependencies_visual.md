# Space Trader Dependency Graph (Visual)

Below is a UML-style diagram (using Mermaid syntax) that visually represents the relationships and dependencies between the major classes, data files, and UI in the Space Trader TUI project.

```mermaid
classDiagram
    class GameState {
        +Player player
        +List~StarSystem~ galaxy
        +StarSystem current_system
        +List~Event~ events
    }
    class Player {
        +Ship ship
        +Dict inventory
        +List~Mercenary~ mercenaries
        +Dict skills
    }
    class Ship {
        +Dict equipment
        +int cargo_bays
        +int weapon_slots
        +int shield_slots
        +int gadget_slots
    }
    class Mercenary {
        +Dict skills
        +int wage
    }
    class StarSystem {
        +Market market
        +String name
        +String tech_level
        +String political_system
        +List resources
        +List events
    }
    class Market {
        +Dict~Good~ goods
        +Dict prices
    }
    class Good {
        +String name
        +int base_price
        +bool legality
        +String min_tech
        +String max_tech
    }
    class Encounter {
        +String type
        +Dict details
    }
    class UI_Screens

    %% Relationships
    GameState --> Player
    GameState --> StarSystem : contains
    Player --> Ship
    Player --> Mercenary : has
    Ship --> Good : equipment uses
    StarSystem --> Market
    Market --> Good
    GameState --> Encounter : triggers
    UI_Screens --> GameState : interacts
    UI_Screens --> Player : displays
    UI_Screens --> StarSystem : displays
    UI_Screens --> Market : displays
    Encounter --> Player : involves
    Encounter --> Ship : involves
    Encounter --> StarSystem : occurs in

    %% Data files
    class systems_json
    class goods_json
    class ships_json
    class equipment_json
    class political_systems_json
    class tech_levels_json
    StarSystem .. systems_json : loaded from
    Good .. goods_json : loaded from
    Ship .. ships_json : loaded from
    Ship .. equipment_json : equipped from
    StarSystem .. political_systems_json : type from
    StarSystem .. tech_levels_json : type from
```

---

**How to view:**
- Paste this Mermaid code block into a Markdown editor that supports Mermaid (e.g., VSCode with the Markdown Preview Mermaid plugin, GitHub, or https://mermaid.live/).
- The diagram will render all major classes, UI, and data file dependencies for easy review.

[Back to Dependencies](space_trader_dependencies.md)
