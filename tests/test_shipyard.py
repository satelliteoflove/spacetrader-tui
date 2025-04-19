import pytest
from player import Player
from world import StarSystem
from ship import Ship
from shipyard import Shipyard

class DummyShip(Ship):
    def __init__(self, type, price=10000, weapon_slots=1, shield_slots=1, gadget_slots=1, min_tech="Industrial"):
        super().__init__()
        self.type = type
        self.price = price
        self.weapon_slots = weapon_slots
        self.shield_slots = shield_slots
        self.gadget_slots = gadget_slots
        self.min_tech = min_tech

class DummyEquipment:
    def __init__(self, name, eq_type, price=1000, min_tech="Industrial"):
        self.name = name
        self.type = eq_type
        self.price = price
        self.min_tech = min_tech

def make_system(tech_level="Industrial"):
    return StarSystem(name="Test", tech_level=tech_level)

def test_buy_ship_success():
    sys = make_system()
    ships = [DummyShip("Gnat", price=10000), DummyShip("Firefly", price=20000)]
    yard = Shipyard(sys, ships=ships, equipment=[])
    player = Player()
    player.ship = ships[0]
    player.credits = 15000
    ok, msg = yard.buy_ship(player, "Firefly")
    assert ok
    assert "Bought Firefly" in msg
    assert player.credits == 5000
    assert player.ship.type == "Firefly"

def test_buy_ship_insufficient_credits():
    sys = make_system()
    ships = [DummyShip("Gnat", price=10000), DummyShip("Firefly", price=30000)]
    yard = Shipyard(sys, ships=ships, equipment=[])
    player = Player()
    player.ship = ships[0]
    player.credits = 10000
    ok, msg = yard.buy_ship(player, "Firefly")
    assert not ok
    assert "Not enough credits" in msg
    assert player.ship.type == "Gnat"

def test_buy_equipment_success():
    sys = make_system()
    ships = [DummyShip("Gnat", price=10000, weapon_slots=1)]
    equip = [DummyEquipment("Pulse Laser", "weapon", price=5000)]
    yard = Shipyard(sys, ships=ships, equipment=equip)
    player = Player()
    player.ship = ships[0]
    player.credits = 6000
    ok, msg = yard.buy_equipment(player, "Pulse Laser")
    assert ok
    assert "Bought and installed Pulse Laser" in msg
    assert player.credits == 1000
    assert "Pulse Laser" in player.ship.equipment["weapons"]

def test_buy_equipment_no_slot():
    sys = make_system()
    ships = [DummyShip("Gnat", price=10000, weapon_slots=0)]
    equip = [DummyEquipment("Pulse Laser", "weapon", price=5000)]
    yard = Shipyard(sys, ships=ships, equipment=equip)
    player = Player()
    player.ship = ships[0]
    player.credits = 6000
    ok, msg = yard.buy_equipment(player, "Pulse Laser")
    assert not ok
    assert "No weapon slots" in msg
    assert player.credits == 6000

def test_remove_equipment():
    sys = make_system()
    ships = [DummyShip("Gnat", price=10000, weapon_slots=1)]
    equip = [DummyEquipment("Pulse Laser", "weapon", price=5000)]
    yard = Shipyard(sys, ships=ships, equipment=equip)
    player = Player()
    player.ship = ships[0]
    player.ship.equipment["weapons"].append("Pulse Laser")
    ok, msg = yard.remove_equipment(player, "weapon", "Pulse Laser")
    assert ok
    assert "Removed Pulse Laser" in msg
    assert "Pulse Laser" not in player.ship.equipment["weapons"]

def test_repair_ship_success():
    sys = make_system()
    ships = [DummyShip("Gnat", price=10000)]
    yard = Shipyard(sys, ships=ships, equipment=[])
    player = Player()
    player.ship = ships[0]
    player.ship.hull = 50
    player.ship.max_hull = 100
    player.credits = 1000
    ok, msg = yard.repair_ship(player)
    assert ok
    assert player.ship.hull == 100
    assert player.credits == 500
    assert "Ship repaired" in msg

def test_repair_ship_no_damage():
    sys = make_system()
    ships = [DummyShip("Gnat", price=10000)]
    yard = Shipyard(sys, ships=ships, equipment=[])
    player = Player()
    player.ship = ships[0]
    player.ship.hull = 100
    player.ship.max_hull = 100
    player.credits = 1000
    ok, msg = yard.repair_ship(player)
    assert not ok
    assert "already fully repaired" in msg
    assert player.credits == 1000
