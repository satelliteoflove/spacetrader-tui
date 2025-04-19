import pytest
from market import Good, Market
from world import StarSystem
from player import Player
from ship import Ship

# Sample goods for testing
goods_list = [
    Good(name="Water", base_price=30, legality=True, min_tech="Pre-agricultural", max_tech="Hi-tech", variance=4),
    Good(name="Firearms", base_price=1250, legality=False, min_tech="Renaissance", max_tech="Hi-tech", variance=100),
]

def make_system(tech_level="Industrial", resources=None, political_system="Democracy", events=None):
    return StarSystem(
        name="TestSystem",
        tech_level=tech_level,
        political_system=political_system,
        resources=resources or [],
        events=events or [],
        goods_list=goods_list
    )

def test_market_price_generation():
    system = make_system()
    market = system.market
    assert "Water" in market.prices
    assert "Firearms" in market.prices
    assert market.prices["Water"] > 0
    assert market.prices["Firearms"] > 0

def test_buy_success():
    system = make_system()
    player = Player()
    player.credits = 10000
    player.ship.cargo_bays = 10
    market = system.market
    success, msg = player.buy_good(market, "Water", 5)
    assert success
    assert player.inventory["Water"] == 5
    assert player.credits < 10000

def test_buy_insufficient_credits():
    system = make_system()
    player = Player()
    player.credits = 10
    market = system.market
    success, msg = player.buy_good(market, "Water", 1)
    assert not success
    assert "Not enough credits" in msg

def test_buy_insufficient_cargo():
    system = make_system()
    player = Player()
    player.credits = 10000
    player.ship.cargo_bays = 2
    market = system.market
    success, msg = player.buy_good(market, "Water", 5)
    assert not success
    assert "Not enough cargo space" in msg

def test_sell_success():
    system = make_system()
    player = Player()
    player.credits = 1000
    player.ship.cargo_bays = 10
    player.inventory["Water"] = 5
    market = system.market
    success, msg = player.sell_good(market, "Water", 3)
    assert success
    assert player.inventory["Water"] == 2
    assert player.credits > 1000

def test_sell_not_enough_goods():
    system = make_system()
    player = Player()
    player.credits = 1000
    player.ship.cargo_bays = 10
    player.inventory["Water"] = 1
    market = system.market
    success, msg = player.sell_good(market, "Water", 5)
    assert not success
    assert "Not enough goods" in msg

def test_illegal_good_buy():
    system = make_system(political_system="Dictatorship")
    player = Player()
    player.credits = 20000
    player.ship.cargo_bays = 10
    market = system.market
    success, msg = player.buy_good(market, "Firearms", 1)
    assert success  # Buying is allowed, but could be flagged for police in future
    assert player.inventory["Firearms"] == 1
    assert player.credits < 20000

def test_market_only_has_tech_goods():
    # Firearms not available in Pre-agricultural
    system = make_system(tech_level="Pre-agricultural")
    market = system.market
    assert "Water" in market.goods
    assert "Firearms" not in market.goods
