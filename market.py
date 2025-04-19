class Good:
    def __init__(self, name, base_price, legality, min_tech, max_tech, variance=None, price_increase_event=None, price_decrease_event=None, expensive_resource=None, cheap_resource=None):
        self.name = name
        self.base_price = base_price
        self.legality = legality
        self.min_tech = min_tech
        self.max_tech = max_tech
        self.variance = variance
        self.price_increase_event = price_increase_event
        self.price_decrease_event = price_decrease_event
        self.expensive_resource = expensive_resource
        self.cheap_resource = cheap_resource

    def to_dict(self):
        return self.__dict__

    def from_dict(self, data):
        self.__dict__.update(data)
        return self

import random

class Market:
    """
    Market holds available goods and their prices for a given star system.
    Prices are generated based on system attributes and events.
    Handles buy/sell transactions.
    """
    def __init__(self, goods_list=None, system=None, events=None):
        """
        goods_list: list of Good objects (from goods.json)
        system: StarSystem instance (for context)
        events: list of current events (optional)
        """
        self.goods = {}
        self.prices = {}
        if goods_list and system:
            self._generate_market(goods_list, system, events or [])

    def _generate_market(self, goods_list, system, events):
        """
        Populate self.goods and self.prices for this system.
        """
        for good in goods_list:
            # Only include goods available at this tech level
            if self._tech_level_ok(system.tech_level, good.min_tech, good.max_tech):
                self.goods[good.name] = good
                self.prices[good.name] = self._calculate_price(good, system, events)

    def _tech_level_ok(self, sys_level, min_tech, max_tech):
        tech_levels = [
            "Pre-agricultural", "Agricultural", "Medieval", "Renaissance",
            "Early Industrial", "Industrial", "Post-industrial", "Hi-tech"
        ]
        try:
            sys_idx = tech_levels.index(sys_level)
            min_idx = tech_levels.index(min_tech)
            max_idx = tech_levels.index(max_tech)
            return min_idx <= sys_idx <= max_idx
        except Exception:
            return False

    def _calculate_price(self, good, system, events):
        """
        Calculate price for a good based on system, events, and randomness.
        """
        price = good.base_price
        # Tech level modifier
        tech_mod = 1.0
        tech_levels = [
            "Pre-agricultural", "Agricultural", "Medieval", "Renaissance",
            "Early Industrial", "Industrial", "Post-industrial", "Hi-tech"
        ]
        try:
            sys_idx = tech_levels.index(system.tech_level)
            min_idx = tech_levels.index(good.min_tech)
            max_idx = tech_levels.index(good.max_tech)
            if sys_idx == min_idx:
                tech_mod -= 0.2  # Cheaper if min tech
            elif sys_idx == max_idx:
                tech_mod += 0.2  # More expensive if max tech
        except Exception:
            pass
        price = int(price * tech_mod)
        # Political system modifier (stub: can expand)
        if system.political_system in ["Dictatorship", "Feudal State"] and not good.legality:
            price = int(price * 1.5)  # Illegal goods expensive in strict govts
        # Resource modifier
        if hasattr(system, 'resources') and good.expensive_resource and good.expensive_resource in system.resources:
            price = int(price * 1.5)
        if hasattr(system, 'resources') and good.cheap_resource and good.cheap_resource in system.resources:
            price = int(price * 0.7)
        # Event modifier
        for event in events:
            if event == good.price_increase_event:
                price = int(price * 1.5)
            if event == good.price_decrease_event:
                price = int(price * 0.7)
        # Random fluctuation
        variance = good.variance or 10
        price = int(price * (1 + random.uniform(-variance/100, variance/100)))
        return max(1, price)

    def buy(self, player, good_name, quantity):
        """
        Player buys quantity of good_name from the market.
        Checks cargo, credits, legality.
        Returns (success, message)
        """
        if good_name not in self.goods:
            return False, "Good not available in this market."
        good = self.goods[good_name]
        price = self.prices[good_name] * quantity
        # Check credits
        if player.credits < price:
            return False, "Not enough credits."
        # Check cargo space
        current_cargo = sum(player.inventory.values())
        if current_cargo + quantity > player.ship.cargo_bays:
            return False, "Not enough cargo space."
        # Check legality (stub: could add police risk)
        if not good.legality:
            # Here we could add a flag for police inspection
            pass
        # Transaction
        player.credits -= price
        player.inventory[good_name] = player.inventory.get(good_name, 0) + quantity
        return True, f"Bought {quantity} {good_name} for {price} credits."

    def sell(self, player, good_name, quantity):
        """
        Player sells quantity of good_name to the market.
        Checks inventory.
        Returns (success, message)
        """
        if good_name not in player.inventory or player.inventory[good_name] < quantity:
            return False, "Not enough goods to sell."
        if good_name not in self.goods:
            return False, "Market does not buy this good."
        price = self.prices[good_name] * quantity
        player.credits += price
        player.inventory[good_name] -= quantity
        if player.inventory[good_name] == 0:
            del player.inventory[good_name]
        return True, f"Sold {quantity} {good_name} for {price} credits."

    def to_dict(self):
        return {
            'goods': {k: v.to_dict() for k, v in self.goods.items()},
            'prices': self.prices,
        }

    def from_dict(self, data):
        self.goods = {k: Good(**v) for k, v in data['goods'].items()}
        self.prices = data['prices']
        return self
