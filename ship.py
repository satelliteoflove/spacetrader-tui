class Ship:
    """
    Represents a player's ship, with attributes for travel, combat, cargo, and equipment.
    """
    def __init__(self):
        self.type = "Gnat"
        self.hull = 100
        self.max_hull = 100
        self.cargo_bays = 15
        self.weapon_slots = 1
        self.shield_slots = 0
        self.gadget_slots = 1
        self.crew_quarters = 1
        self.range = 14  # max travel distance per tank (parsecs)
        self.fuel = 14   # start full; max = range
        # Equipment dict: {'weapons': [], 'shields': [], 'gadgets': []}
        self.equipment = {'weapons': [], 'shields': [], 'gadgets': []}

    def can_travel(self, distance):
        """Return True if the ship has enough fuel to travel the given distance."""
        return self.fuel >= distance

    def travel(self, distance):
        """Deduct fuel for travel. Return True if successful, False if not enough fuel."""
        if self.can_travel(distance):
            self.fuel -= distance
            return True
        return False

    def install_equipment(self, eq_type, eq_name):
        """Install equipment of type (weapon/shield/gadget) if slot available."""
        slots = self._get_slot_count(eq_type)
        eq_list = self.equipment[eq_type + 's']
        if len(eq_list) >= slots:
            return False, f"No {eq_type} slots available."
        eq_list.append(eq_name)
        self.apply_gadget_effects()
        return True, f"Installed {eq_name} as {eq_type}."

    def remove_equipment(self, eq_type, eq_name):
        """Remove equipment of type (weapon/shield/gadget) if present."""
        eq_list = self.equipment[eq_type + 's']
        if eq_name in eq_list:
            eq_list.remove(eq_name)
            self.apply_gadget_effects()
            return True, f"Removed {eq_name} from {eq_type}s."
        return False, f"{eq_name} not installed."

    def _get_slot_count(self, eq_type):
        if eq_type == 'weapon':
            return self.weapon_slots
        if eq_type == 'shield':
            return self.shield_slots
        if eq_type == 'gadget':
            return self.gadget_slots
        return 0

    def repair(self):
        """Restore hull to max_hull."""
        self.hull = self.max_hull
        return True, "Ship fully repaired."

    def apply_gadget_effects(self):
        """
        Apply gadget effects to ship stats (cargo_bays, skill bonuses, etc).
        Should be called after installing/removing gadgets or loading from dict.
        """
        try:
            from loaders import load_equipment
            equipment_db = load_equipment('data/equipment.json')
        except Exception:
            equipment_db = []
        eq_lookup = {e['name']: e for e in equipment_db}

        # Reset dynamic bonuses
        self._gadget_cargo_bonus = 0
        self._gadget_pilot_bonus = 0
        self._gadget_fighter_bonus = 0
        self._gadget_engineer_bonus = 0
        self._has_cloaking = False
        self._has_auto_repair = False

        for name in self.equipment.get('gadgets', []):
            eq = eq_lookup.get(name, {})
            if name == 'Extra Cargo Bays':
                self._gadget_cargo_bonus += eq.get('bonus', 0)
            elif name == 'Navigation System':
                self._gadget_pilot_bonus += 1
            elif name == 'Targeting System':
                self._gadget_fighter_bonus += 1
            elif name == 'Auto-Repair System':
                self._has_auto_repair = True
                self._gadget_engineer_bonus += 1
            elif name == 'Cloaking Device':
                self._has_cloaking = True
                self._gadget_pilot_bonus += 1

    @property
    def cargo_bays_total(self):
        """Total cargo bays including gadget bonuses."""
        return self.cargo_bays + getattr(self, '_gadget_cargo_bonus', 0)

    @property
    def has_cloaking(self):
        return getattr(self, '_has_cloaking', False)

    @property
    def has_auto_repair(self):
        return getattr(self, '_has_auto_repair', False)

    @property
    def pilot_skill_bonus(self):
        return getattr(self, '_gadget_pilot_bonus', 0)

    @property
    def fighter_skill_bonus(self):
        return getattr(self, '_gadget_fighter_bonus', 0)

    @property
    def engineer_skill_bonus(self):
        return getattr(self, '_gadget_engineer_bonus', 0)

    def to_dict(self):
        return self.__dict__

    def from_dict(self, data):
        self.__dict__.update(data)
        return self

class Mercenary:
    def __init__(self):
        self.name = "Merc"
        self.skills = {'pilot': 1, 'fighter': 1, 'trader': 1, 'engineer': 1}
        self.wage = 100

    def to_dict(self):
        return self.__dict__

    def from_dict(self, data):
        self.__dict__.update(data)
        return self
