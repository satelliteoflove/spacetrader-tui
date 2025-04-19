import pytest
from game import GameState
from world import StarSystem

# Helper for custom system creation
class DummySystem(StarSystem):
    def __init__(self, name, x, y):
        super().__init__(name=name, x=x, y=y)


def test_distance_calculation():
    sys1 = DummySystem("A", 0, 0)
    sys2 = DummySystem("B", 3, 4)
    assert sys1.distance_to(sys2) == 5.0

def test_get_reachable_systems():
    gs = GameState()
    # Place current system at (0,0), others at (3,4) and (20,0)
    sys0 = DummySystem("A", 0, 0)
    sys1 = DummySystem("B", 3, 4)
    sys2 = DummySystem("C", 20, 0)
    gs.systems = [sys0, sys1, sys2]
    gs.current_system = sys0
    gs.player.ship.fuel = 10
    gs.player.ship.range = 10
    reachable = gs.get_reachable_systems()
    names = [s.name for s, d in reachable]
    assert "B" in names
    assert "C" not in names

def test_travel_success():
    gs = GameState()
    sys0 = DummySystem("A", 0, 0)
    sys1 = DummySystem("B", 3, 4)
    gs.systems = [sys0, sys1]
    gs.current_system = sys0
    gs.player.ship.fuel = 10
    gs.player.ship.range = 10
    success, msg = gs.travel_to_system(sys1)
    assert success
    assert gs.current_system is sys1
    assert gs.player.ship.fuel == 5  # 10 - 5
    assert "Traveled to B" in msg

def test_travel_insufficient_fuel():
    gs = GameState()
    sys0 = DummySystem("A", 0, 0)
    sys1 = DummySystem("B", 30, 40)  # distance = 50
    gs.systems = [sys0, sys1]
    gs.current_system = sys0
    gs.player.ship.fuel = 10
    gs.player.ship.range = 10
    success, msg = gs.travel_to_system(sys1)
    assert not success
    assert "Not enough fuel" in msg
    assert gs.current_system is sys0

def test_travel_invalid_destination():
    gs = GameState()
    sys0 = DummySystem("A", 0, 0)
    gs.systems = [sys0]
    gs.current_system = sys0
    success, msg = gs.travel_to_system(None)
    assert not success
    assert "Invalid current or destination" in msg
