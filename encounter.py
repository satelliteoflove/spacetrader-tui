# Placeholder for encounter logic (pirates, traders, police, combat)

import random

import random

class Encounter:
    """
    Represents a random encounter (pirate, trader, police) during travel.
    Choices and outcomes depend on player/ship/crew stats.
    """
    def __init__(self, encounter_type=None):
        self.type = encounter_type  # 'pirate', 'trader', 'police'
        self.details = {}

    @staticmethod
    def random_encounter(player, system):
        """
        Randomly select an encounter type based on system, player, and cargo.
        Returns an Encounter instance or None (no encounter).
        """
        roll = random.random()
        has_illegal = any(
            not g.legality and qty > 0
            for gname, qty in player.inventory.items()
            for g in system.market.goods.values() if g.name == gname
        )
        police_chance = 0.2 + (0.2 if has_illegal else 0)
        pirate_chance = 0.15
        trader_chance = 0.10
        if roll < police_chance:
            return Encounter('police')
        elif roll < police_chance + pirate_chance:
            return Encounter('pirate')
        elif roll < police_chance + pirate_chance + trader_chance:
            return Encounter('trader')
        else:
            return None  # No encounter

    def resolve(self, player, system, choice=None, **kwargs):
        """
        Resolve the encounter. Returns a message.
        Player choice can be passed as a parameter (for tests/UI).
        """
        if self.type == 'police':
            return self._resolve_police(player, system, choice, **kwargs)
        elif self.type == 'pirate':
            return self._resolve_pirate(player, system, choice, **kwargs)
        elif self.type == 'trader':
            return self._resolve_trader(player, system, choice, **kwargs)
        else:
            return "No encounter."

    # --- Police Encounter ---
    def _resolve_police(self, player, system, choice=None, **kwargs):
        # Choices: 'comply', 'bribe', 'flee'
        if not choice:
            choice = random.choice(['comply', 'bribe', 'flee'])
        if choice == 'comply':
            illegal_goods = [gname for gname, qty in player.inventory.items()
                             for g in system.market.goods.values()
                             if g.name == gname and not g.legality and qty > 0]
            if illegal_goods:
                total_fine = 0
                for gname in illegal_goods:
                    qty = player.inventory[gname]
                    fine = 500 * qty
                    total_fine += fine
                    del player.inventory[gname]
                player.credits = max(0, player.credits - total_fine)
                player.police_record += 1
                return f"Police inspected your ship! Illegal goods confiscated. Fined {total_fine} credits. Police record worsened."
            else:
                return "Police inspected your ship but found nothing illegal."
        elif choice == 'bribe':
            bribe_amount = kwargs.get('bribe_amount', 1000)
            trader_skill = player.skills.get('trader', 1)
            success_chance = 0.2 + 0.1 * trader_skill + min(bribe_amount / 5000, 0.5)
            if player.credits < bribe_amount:
                return "Bribe failed: not enough credits."
            player.credits -= bribe_amount
            if random.random() < success_chance:
                return f"Bribe of {bribe_amount} credits accepted. Police let you go."
            else:
                player.police_record += 1
                return f"Bribe failed! Police record worsened."
        elif choice == 'flee':
            pilot_skill = player.skills.get('pilot', 1)
            flee_chance = 0.3 + 0.1 * pilot_skill
            if random.random() < flee_chance:
                return "You successfully fled from the police!"
            else:
                player.police_record += 1
                return "Failed to flee. Police record worsened."
        else:
            return "Police encounter: invalid choice."

    # --- Pirate Encounter ---
    def _resolve_pirate(self, player, system, choice=None, **kwargs):
        # Choices: 'fight', 'flee', 'surrender'
        if not choice:
            choice = random.choice(['fight', 'flee', 'surrender'])

        # --- Load equipment data ---
        try:
            from loaders import load_equipment
            equipment_db = load_equipment('data/equipment.json')
        except Exception:
            equipment_db = []
        eq_lookup = {e['name']: e for e in equipment_db}

        # Helper: sum stat for equipped items
        def sum_equipment_stat(eq_names, stat):
            return sum(eq_lookup.get(name, {}).get(stat, 0) for name in eq_names)

        # Helper: check for gadget
        def has_gadget(gadget_name):
            return gadget_name in player.ship.equipment.get('gadgets', [])

        if choice == 'fight':
            fighter_skill = player.skills.get('fighter', 1)
            crew_fighter = sum(m.skills.get('fighter', 0) for m in getattr(player, 'crew', []))
            weapon_names = player.ship.equipment.get('weapons', [])
            weapon_power = sum_equipment_stat(weapon_names, 'power')
            power = fighter_skill + crew_fighter + weapon_power
            pirate_power = random.randint(10, 50)  # scale up since weapon_power can be large
            if power >= pirate_power:
                loot = random.randint(500, 2000)
                player.credits += loot
                # Auto-repair gadget stub (to be implemented)
                # if has_gadget('Auto-Repair System'):
                #     ...
                return f"You fought off the pirates and looted {loot} credits! (Your combat power: {power}, Pirate: {pirate_power})"
            else:
                # Calculate shield protection
                shield_names = player.ship.equipment.get('shields', [])
                shield_protection = sum_equipment_stat(shield_names, 'protection')
                raw_damage = random.randint(10, 40)
                damage = max(0, raw_damage - shield_protection)
                player.ship.hull = max(0, player.ship.hull - damage)
                lost = min(player.credits, random.randint(100, 500))
                player.credits -= lost
                # Auto-repair gadget stub (to be implemented)
                # if has_gadget('Auto-Repair System'):
                #     ...
                return f"You lost the fight! Ship took {damage} damage (shields absorbed {shield_protection}). Lost {lost} credits."
        elif choice == 'flee':
            # Cloaking device stub (to be implemented)
            # if has_gadget('Cloaking Device'):
            #     ...
            pilot_skill = player.skills.get('pilot', 1)
            crew_pilot = sum(m.skills.get('pilot', 0) for m in getattr(player, 'crew', []))
            flee_chance = 0.3 + 0.1 * (pilot_skill + crew_pilot)
            if random.random() < flee_chance:
                return "You successfully fled from the pirates!"
            else:
                shield_names = player.ship.equipment.get('shields', [])
                shield_protection = sum_equipment_stat(shield_names, 'protection')
                raw_damage = random.randint(5, 20)
                damage = max(0, raw_damage - shield_protection)
                player.ship.hull = max(0, player.ship.hull - damage)
                lost = min(player.credits, random.randint(50, 300))
                player.credits -= lost
                return f"Failed to flee. Ship took {damage} damage (shields absorbed {shield_protection}). Lost {lost} credits."
        elif choice == 'surrender':
            lost = min(player.credits, random.randint(200, 1000))
            player.credits -= lost
            cargo_lost = []
            for g in list(player.inventory.keys()):
                if random.random() < 0.5:
                    cargo_lost.append(g)
                    del player.inventory[g]
            return f"You surrendered. Lost {lost} credits and cargo: {', '.join(cargo_lost) if cargo_lost else 'none'}."
        else:
            return "Pirate encounter: invalid choice."

        elif choice == 'flee':
            pilot_skill = player.skills.get('pilot', 1)
            crew_pilot = sum(m.skills.get('pilot', 0) for m in getattr(player, 'crew', []))
            flee_chance = 0.3 + 0.1 * (pilot_skill + crew_pilot)
            if random.random() < flee_chance:
                return "You successfully fled from the pirates!"
            else:
                damage = random.randint(5, 20)
                player.ship.hull = max(0, player.ship.hull - damage)
                lost = min(player.credits, random.randint(50, 300))
                player.credits -= lost
                return f"Failed to flee. Ship took {damage} damage and lost {lost} credits."
        elif choice == 'surrender':
            lost = min(player.credits, random.randint(200, 1000))
            player.credits -= lost
            cargo_lost = []
            for g in list(player.inventory.keys()):
                if random.random() < 0.5:
                    cargo_lost.append(g)
                    del player.inventory[g]
            return f"You surrendered. Lost {lost} credits and cargo: {', '.join(cargo_lost) if cargo_lost else 'none'}."
        else:
            return "Pirate encounter: invalid choice."

    # --- Trader Encounter ---
    def _resolve_trader(self, player, system, choice=None, **kwargs):
        # Choices: 'trade', 'decline'
        if not choice:
            choice = random.choice(['trade', 'decline'])
        if choice == 'trade':
            # Offer to buy or sell a random good
            goods = list(system.market.goods.values())
            if not goods:
                return "Trader has nothing to trade."
            good = random.choice(goods)
            price = system.market.prices[good.name]
            trader_skill = player.skills.get('trader', 1)
            # Trader offers a price with some negotiation
            offer_price = int(price * (0.9 + 0.02 * trader_skill))
            # Player buys 1 unit if they can afford and have space
            current_cargo = sum(player.inventory.values())
            if player.credits >= offer_price and current_cargo < player.ship.cargo_bays:
                player.credits -= offer_price
                player.inventory[good.name] = player.inventory.get(good.name, 0) + 1
                return f"Traded with the trader: bought 1 {good.name} for {offer_price} credits."
            else:
                return "Could not trade: not enough credits or cargo space."
        elif choice == 'decline':
            return "You declined to trade with the trader."
        else:
            return "Trader encounter: invalid choice."

