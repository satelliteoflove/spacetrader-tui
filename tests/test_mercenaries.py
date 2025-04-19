import pytest
from player import Player
from ship import Ship
from mercenary import Mercenary
from merc_market import get_available_mercenaries

def make_player_with_ship(crew_quarters=2, credits=5000):
    player = Player()
    ship = Ship()
    ship.crew_quarters = crew_quarters
    player.ship = ship
    player.credits = credits
    return player

def test_hire_mercenary_success():
    player = make_player_with_ship(crew_quarters=2, credits=5000)
    merc = Mercenary("Test Merc", "fighter", {"fighter": 3}, 1000)
    ok, msg = player.hire_crew(merc)
    assert ok
    assert "Hired" in msg
    assert merc in player.crew
    assert player.credits == 4000

def test_hire_mercenary_no_quarters():
    player = make_player_with_ship(crew_quarters=1, credits=5000)
    merc1 = Mercenary("Merc1", "pilot", {"pilot": 2}, 1000)
    merc2 = Mercenary("Merc2", "fighter", {"fighter": 2}, 1000)
    player.hire_crew(merc1)
    ok, msg = player.hire_crew(merc2)
    assert not ok
    assert "No available crew quarters" in msg
    assert merc2 not in player.crew

def test_hire_mercenary_insufficient_credits():
    player = make_player_with_ship(crew_quarters=2, credits=500)
    merc = Mercenary("Test Merc", "trader", {"trader": 2}, 1000)
    ok, msg = player.hire_crew(merc)
    assert not ok
    assert "Not enough credits" in msg
    assert merc not in player.crew

def test_fire_mercenary():
    player = make_player_with_ship(crew_quarters=2, credits=5000)
    merc = Mercenary("Test Merc", "engineer", {"engineer": 3}, 1000)
    player.hire_crew(merc)
    ok, msg = player.fire_crew("Test Merc")
    assert ok
    assert "Fired Test Merc" in msg
    assert merc not in player.crew

def test_fire_mercenary_not_found():
    player = make_player_with_ship(crew_quarters=2, credits=5000)
    ok, msg = player.fire_crew("Ghost Merc")
    assert not ok
    assert "No mercenary named Ghost Merc found" in msg

def test_get_available_mercenaries():
    mercs = get_available_mercenaries(max_count=3)
    assert len(mercs) == 3
    assert all(isinstance(m, Mercenary) for m in mercs)
