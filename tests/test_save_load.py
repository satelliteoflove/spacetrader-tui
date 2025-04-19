import os
import tempfile
from player import Player
from world import StarSystem
from market import Good
from save_load import save_game, load_game

goods_list = [
    Good(name="Water", base_price=30, legality=True, min_tech="Pre-agricultural", max_tech="Hi-tech"),
    Good(name="Narcotics", base_price=3500, legality=False, min_tech="Industrial", max_tech="Hi-tech"),
]

def make_system():
    sys = StarSystem(name="Test", x=0, y=0, goods_list=goods_list)
    return sys

def test_save_and_load_roundtrip():
    player = Player()
    player.name = "Tester"
    player.credits = 1234
    player.inventory["Water"] = 2
    player.skills['pilot'] = 5
    # Add a fake current_system attribute for meta
    player.current_system = "Test"
    systems = [make_system()]
    meta = {'turn': 42, 'time': '2025-04-17T23:45:00-04:00', 'current_system': 'Test', 'random_seed': 12345}
    with tempfile.NamedTemporaryFile(delete=False) as tf:
        fname = tf.name
    try:
        save_game(fname, player, systems, meta)
        loaded_player, loaded_systems, loaded_meta = load_game(fname, goods_list=goods_list)
        assert loaded_player.name == "Tester"
        assert loaded_player.credits == 1234
        assert loaded_player.inventory["Water"] == 2
        assert loaded_player.skills['pilot'] == 5
        assert loaded_systems[0].name == systems[0].name
        assert loaded_meta['turn'] == 42
        assert loaded_meta['time'] == '2025-04-17T23:45:00-04:00'
        assert loaded_meta['current_system'] == 'Test'
        assert loaded_meta['random_seed'] == 12345
    finally:
        os.remove(fname)

def test_save_load_empty_inventory_and_crew():
    player = Player()
    player.crew = []
    player.inventory = {}
    systems = [make_system()]
    meta = {'current_system': 'Test'}
    with tempfile.NamedTemporaryFile(delete=False) as tf:
        fname = tf.name
    try:
        save_game(fname, player, systems, meta)
        loaded_player, loaded_systems, loaded_meta = load_game(fname, goods_list=goods_list)
        assert loaded_player.inventory == {}
        assert loaded_player.crew == []
    finally:
        os.remove(fname)

def test_save_load_multiple_systems():
    player = Player()
    systems = [make_system(), make_system()]
    systems[1].name = "Alpha"
    meta = {'current_system': 'Alpha'}
    with tempfile.NamedTemporaryFile(delete=False) as tf:
        fname = tf.name
    try:
        save_game(fname, player, systems, meta)
        _, loaded_systems, loaded_meta = load_game(fname, goods_list=goods_list)
        assert len(loaded_systems) == 2
        assert loaded_systems[1].name == "Alpha"
        assert loaded_meta['current_system'] == 'Alpha'
    finally:
        os.remove(fname)

def test_save_load_complex_player_and_ship():
    from mercenary import Mercenary
    player = Player()
    player.name = "Ace"
    player.credits = 9999
    player.crew = [Mercenary("Merc1", "pilot", {"pilot": 3}, 100, status="active")]
    player.ship.hull = 42
    player.ship.equipment['weapons'] = ["Laser", "Missile"]
    player.inventory = {"Narcotics": 1, "Water": 5}
    systems = [make_system()]
    meta = {'current_system': 'Test'}
    with tempfile.NamedTemporaryFile(delete=False) as tf:
        fname = tf.name
    try:
        save_game(fname, player, systems, meta)
        loaded_player, _, _ = load_game(fname, goods_list=goods_list)
        assert loaded_player.name == "Ace"
        assert loaded_player.credits == 9999
        assert isinstance(loaded_player.crew, list)
        assert loaded_player.ship.hull == 42
        assert "Laser" in loaded_player.ship.equipment['weapons']
        assert loaded_player.inventory["Narcotics"] == 1
    finally:
        os.remove(fname)

def test_save_load_market_edge_cases():
    from market import Market
    player = Player()
    systems = [make_system()]
    # Remove all goods from market
    systems[0].market.goods = {}
    systems[0].market.prices = {}
    meta = {'current_system': 'Test'}
    with tempfile.NamedTemporaryFile(delete=False) as tf:
        fname = tf.name
    try:
        save_game(fname, player, systems, meta)
        _, loaded_systems, _ = load_game(fname, goods_list=goods_list)
        assert loaded_systems[0].market.goods == {}
        assert loaded_systems[0].market.prices == {}
    finally:
        os.remove(fname)

def test_save_load_missing_meta_fields():
    player = Player()
    systems = [make_system()]
    meta = {}  # No fields
    with tempfile.NamedTemporaryFile(delete=False) as tf:
        fname = tf.name
    try:
        save_game(fname, player, systems, meta)
        _, _, loaded_meta = load_game(fname, goods_list=goods_list)
        # Should have at least some default fields (auto-populated or empty)
        assert 'turn' in loaded_meta or 'current_system' in loaded_meta or 'time' in loaded_meta
    finally:
        os.remove(fname)

def test_save_load_extra_meta_fields():
    player = Player()
    systems = [make_system()]
    meta = {'current_system': 'Test', 'bonus': 777, 'difficulty': 'Hard'}
    with tempfile.NamedTemporaryFile(delete=False) as tf:
        fname = tf.name
    try:
        save_game(fname, player, systems, meta)
        _, _, loaded_meta = load_game(fname, goods_list=goods_list)
        assert loaded_meta['bonus'] == 777
        assert loaded_meta['difficulty'] == 'Hard'
    finally:
        os.remove(fname)

def test_save_load_corrupted_file():
    import io
    # Write incomplete/corrupted JSON
    with tempfile.NamedTemporaryFile(delete=False, mode='w') as tf:
        tf.write('{"player":')
        fname = tf.name
    try:
        try:
            load_game(fname, goods_list=goods_list)
            assert False, "Should have raised an exception for corrupted file"
        except Exception:
            pass  # Expected
    finally:
        os.remove(fname)
