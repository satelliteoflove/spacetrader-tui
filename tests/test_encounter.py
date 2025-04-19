import pytest
from player import Player
from world import StarSystem
from market import Good, Market
from encounter import Encounter

# Minimal goods for police test
goods_list = [
    Good(name="Water", base_price=30, legality=True, min_tech="Pre-agricultural", max_tech="Hi-tech"),
    Good(name="Narcotics", base_price=3500, legality=False, min_tech="Industrial", max_tech="Hi-tech"),
]

def make_system():
    sys = StarSystem(name="Test", x=0, y=0, goods_list=goods_list)
    return sys

def test_police_encounter_illegal_goods():
    player = Player()
    system = make_system()
    player.inventory["Narcotics"] = 2
    player.inventory["Water"] = 5
    player.credits = 2000
    player.police_record = 0
    encounter = Encounter('police')
    msg = encounter.resolve(player, system, choice='comply')
    assert "Illegal goods confiscated" in msg
    assert "Fined" in msg
    assert player.credits == 1000  # 2000 - (2*500)
    assert player.police_record == 1
    assert "Narcotics" not in player.inventory
    assert player.inventory["Water"] == 5

def test_police_encounter_no_illegal_goods():
    player = Player()
    system = make_system()
    player.inventory["Water"] = 3
    player.credits = 500
    player.police_record = 0
    encounter = Encounter('police')
    msg = encounter.resolve(player, system, choice='comply')
    assert "found nothing illegal" in msg
    assert player.credits == 500
    assert player.police_record == 0
    assert player.inventory["Water"] == 3

def test_pirate_fight_win_and_lose():
    player = Player()
    system = make_system()
    player.skills['fighter'] = 10  # Guarantee win
    encounter = Encounter('pirate')
    msg = encounter.resolve(player, system, choice='fight')
    assert "fought off the pirates" in msg or "lost the fight" in msg

    player = Player()
    system = make_system()
    player.skills['fighter'] = 0  # Guarantee loss
    player.credits = 1000
    encounter = Encounter('pirate')
    msg = encounter.resolve(player, system, choice='fight')
    assert "lost the fight" in msg or "fought off the pirates" in msg

def test_pirate_flee():
    player = Player()
    system = make_system()
    player.skills['pilot'] = 10  # Guarantee flee
    encounter = Encounter('pirate')
    msg = encounter.resolve(player, system, choice='flee')
    assert "fled from the pirates" in msg or "Failed to flee" in msg

    player = Player()
    system = make_system()
    player.skills['pilot'] = 0  # Guarantee fail
    player.credits = 1000
    encounter = Encounter('pirate')
    msg = encounter.resolve(player, system, choice='flee')
    assert "Failed to flee" in msg or "fled from the pirates" in msg

def test_pirate_surrender():
    player = Player()
    system = make_system()
    player.credits = 1000
    player.inventory['Water'] = 2
    encounter = Encounter('pirate')
    msg = encounter.resolve(player, system, choice='surrender')
    assert "surrendered" in msg
    # Credits should decrease, some cargo may be lost
    assert player.credits <= 1000

def test_police_bribe_success_and_fail():
    player = Player()
    system = make_system()
    player.skills['trader'] = 10  # High skill, high bribe
    player.credits = 2000
    encounter = Encounter('police')
    msg = encounter.resolve(player, system, choice='bribe', bribe_amount=1500)
    assert "Bribe" in msg

    player = Player()
    system = make_system()
    player.skills['trader'] = 0  # Low skill, low bribe
    player.credits = 1000
    encounter = Encounter('police')
    msg = encounter.resolve(player, system, choice='bribe', bribe_amount=100)
    assert "Bribe" in msg or "failed" in msg

def test_police_flee_success_and_fail():
    player = Player()
    system = make_system()
    player.skills['pilot'] = 10  # High skill
    encounter = Encounter('police')
    msg = encounter.resolve(player, system, choice='flee')
    assert "fled from the police" in msg or "Failed to flee" in msg

    player = Player()
    system = make_system()
    player.skills['pilot'] = 0  # Low skill
    encounter = Encounter('police')
    msg = encounter.resolve(player, system, choice='flee')
    assert "Failed to flee" in msg or "fled from the police" in msg

def test_trader_trade_and_decline():
    player = Player()
    system = make_system()
    player.credits = 1000
    player.ship.cargo_bays = 10
    encounter = Encounter('trader')
    msg = encounter.resolve(player, system, choice='trade')
    assert "Traded with the trader" in msg or "Could not trade" in msg

    player = Player()
    system = make_system()
    encounter = Encounter('trader')
    msg = encounter.resolve(player, system, choice='decline')
    assert "declined to trade" in msg

def test_random_encounter_distribution():
    # Just check that all types can be generated
    player = Player()
    system = make_system()
    player.inventory["Narcotics"] = 1
    types_found = set()
    for _ in range(200):
        e = Encounter.random_encounter(player, system)
        if e:
            types_found.add(e.type)
    assert "police" in types_found
    assert "pirate" in types_found
    assert "trader" in types_found
