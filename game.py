import json
from player import Player
from world import StarSystem
from loaders import load_star_systems, load_goods, load_equipment, load_ships

class GameState:
    def __init__(self):
        import random
        # Load all game data at startup
        self.systems = load_star_systems('data/systems.json')
        self.goods = load_goods('data/goods.json')
        self.equipment = load_equipment('data/equipment.json')
        self.ships = load_ships('data/ships.json')

        # --- Player Initialization ---
        self.player = Player()
        self.player.name = self._generate_player_name()
        self.player.credits = 1000  # Or randomize for challenge
        self.player.skills = {'pilot': 1, 'fighter': 1, 'trader': 1, 'engineer': 1}  # Could be randomized/assigned later
        self.player.inventory = {}
        self.player.ship = random.choice(self.ships) if self.ships else Player().ship
        self.player.mercenaries = []
        # Assign starting system randomly
        self.current_system = random.choice(self.systems) if self.systems else None
        # Optionally, set player's location to current_system
        self.galaxy = self.systems  # Alias for legacy code, can be refactored
        self.events = []

    def get_reachable_systems(self):
        """
        Return a list of StarSystems within the player's ship fuel range.
        """
        if not self.current_system:
            return []
        ship = self.player.ship
        reachable = []
        for system in self.systems:
            if system is self.current_system:
                continue
            try:
                dist = self.current_system.distance_to(system)
                if ship.can_travel(dist):
                    reachable.append((system, dist))
            except Exception:
                continue
        return reachable

    def travel_to_system(self, destination):
        """
        Attempt to travel to the destination StarSystem.
        Checks fuel, deducts fuel, moves player, and triggers encounter logic.
        Returns (success, combined_message)
        """
        if not self.current_system or not destination:
            return False, "Invalid current or destination system."
        ship = self.player.ship
        try:
            dist = self.current_system.distance_to(destination)
        except Exception:
            return False, "Cannot calculate distance to destination."
        if not ship.can_travel(dist):
            return False, "Not enough fuel to reach destination."
        # Deduct fuel
        ship.travel(dist)
        # Move player
        self.current_system = destination
        # Advance game state (e.g., day, events)
        # TODO: Add day/event advancement logic if needed
        # Trigger encounter
        from encounter import Encounter
        encounter = Encounter.random_encounter(self.player, self.current_system)
        travel_msg = f"Traveled to {destination.name} ({dist:.1f} parsecs). Fuel remaining: {ship.fuel:.1f}."
        if encounter:
            encounter_msg = encounter.resolve(self.player, self.current_system)
            return True, f"{travel_msg}\nEncounter: {encounter.type.capitalize()}\n{encounter_msg}"
        else:
            return True, f"{travel_msg}\nNo encounter."

    def _generate_player_name(self):
        # Placeholder: could prompt user or randomize
        import random
        names = ["Trader", "Captain", "Astra", "Nova", "Orion", "Zephyr", "Vega"]
        return random.choice(names)

    def run(self):
        # Placeholder for main game loop
        print("Welcome to Space Trader (TUI)!")
        # TODO: Launch Textual UI

    def save(self, filename):
        with open(filename, 'w') as f:
            json.dump(self.to_dict(), f, indent=2)

    def load(self, filename):
        with open(filename, 'r') as f:
            data = json.load(f)
            self.from_dict(data)

    def to_dict(self):
        return {
            'player': self.player.to_dict(),
            'galaxy': [s.to_dict() for s in self.galaxy],
            'current_system': self.current_system,
            'events': self.events,
        }

    def from_dict(self, data):
        self.player.from_dict(data['player'])
        self.galaxy = [StarSystem().from_dict(sd) for sd in data['galaxy']]
        self.current_system = data['current_system']
        self.events = data['events']
