from ship import Ship, Mercenary

class Player:
    def __init__(self):
        self.name = "Trader"
        self.credits = 1000
        self.loan_balance = 0  # Amount owed to the bank
        self.skills = {'pilot': 1, 'fighter': 1, 'trader': 1, 'engineer': 1}
        self.reputation = 0
        self.police_record = 0
        self.ship = Ship()
        self.inventory = {}
        self.crew = []  # List of mercenaries/crew

    def can_hire_crew(self):
        """Return True if there is space for more crew."""
        return len(self.crew) < getattr(self.ship, 'crew_quarters', 1)

    def hire_crew(self, merc):
        """Hire a mercenary if there is space and enough credits."""
        if not self.can_hire_crew():
            return False, "No available crew quarters."
        if self.credits < merc.salary:
            return False, "Not enough credits to hire mercenary."
        self.credits -= merc.salary
        self.crew.append(merc)
        return True, f"Hired {merc.name} as {merc.role}."

    def fire_crew(self, merc_name):
        """Fire a mercenary by name."""
        for m in self.crew:
            if m.name == merc_name:
                self.crew.remove(m)
                return True, f"Fired {merc_name}."
        return False, f"No mercenary named {merc_name} found."

    def buy_good(self, market, good_name, quantity):
        """
        Attempt to buy a quantity of good from the given market.
        Returns (success, message)
        """
        return market.buy(self, good_name, quantity)

    def sell_good(self, market, good_name, quantity):
        """
        Attempt to sell a quantity of good to the given market.
        Returns (success, message)
        """
        return market.sell(self, good_name, quantity)

    def to_dict(self):
        return {
            'name': self.name,
            'credits': self.credits,
            'skills': self.skills,
            'reputation': self.reputation,
            'police_record': self.police_record,
            'ship': self.ship.to_dict(),
            'inventory': self.inventory,
            'crew': [m.to_dict() for m in self.crew],
        }

    def from_dict(self, data):
        self.name = data['name']
        self.credits = data['credits']
        self.skills = data['skills']
        self.reputation = data['reputation']
        self.police_record = data['police_record']
        self.ship.from_dict(data['ship'])
        self.inventory = data['inventory']
        from mercenary import Mercenary
        self.crew = [Mercenary.from_dict(md) for md in data.get('crew',[])]

