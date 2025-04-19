from loaders import load_ships, load_equipment

class Shipyard:
    """
    Shipyard provides ship/equipment purchase, sale, and repair services.
    """
    def __init__(self, system, ships=None, equipment=None):
        self.system = system
        # Available ships/equipment filtered by system tech level
        self.ships = ships if ships is not None else load_ships('data/ships.json')
        self.equipment = equipment if equipment is not None else load_equipment('data/equipment.json')

    def available_ships(self):
        """Return list of ships available for purchase (by tech level)."""
        sys_level = self.system.tech_level
        return [s for s in self.ships if self._tech_ok(sys_level, s.min_tech)]

    def available_equipment(self):
        """Return list of equipment available for purchase (by tech level)."""
        sys_level = self.system.tech_level
        return [e for e in self.equipment if self._tech_ok(sys_level, e.min_tech)]

    def buy_ship(self, player, ship_type):
        """Buy a new ship, trading in the old one (if player can afford it)."""
        for s in self.available_ships():
            if s.type == ship_type:
                trade_in = player.ship.price if hasattr(player.ship, 'price') else 0
                price = s.price - trade_in
                if player.credits < price:
                    return False, "Not enough credits to buy this ship."
                player.credits -= price
                player.ship = s
                return True, f"Bought {ship_type} for {price} credits (trade-in applied)."
        return False, "Ship not available."

    def buy_equipment(self, player, eq_name):
        """Buy and install equipment if player can afford it and has slot."""
        for e in self.available_equipment():
            if e.name == eq_name:
                if player.credits < e.price:
                    return False, "Not enough credits to buy equipment."
                eq_type = e.type  # 'weapon', 'shield', 'gadget'
                ok, msg = player.ship.install_equipment(eq_type, eq_name)
                if not ok:
                    return False, msg
                player.credits -= e.price
                return True, f"Bought and installed {eq_name} for {e.price} credits."
        return False, "Equipment not available."

    def remove_equipment(self, player, eq_type, eq_name):
        """Remove equipment from ship."""
        return player.ship.remove_equipment(eq_type, eq_name)

    def repair_ship(self, player):
        """Repair ship hull at a cost (10 credits per missing hull point)."""
        missing = player.ship.max_hull - player.ship.hull
        if missing <= 0:
            return False, "Ship is already fully repaired."
        cost = 10 * missing
        if player.credits < cost:
            return False, "Not enough credits to repair ship."
        player.credits -= cost
        player.ship.repair()
        return True, f"Ship repaired for {cost} credits."

    def _tech_ok(self, sys_level, min_tech):
        tech_levels = [
            "Pre-agricultural", "Agricultural", "Medieval", "Renaissance",
            "Early Industrial", "Industrial", "Post-industrial", "Hi-tech"
        ]
        try:
            sys_idx = tech_levels.index(sys_level)
            min_idx = tech_levels.index(min_tech)
            return sys_idx >= min_idx
        except Exception:
            return False
