import os
import tempfile
import json
import pytest
from loaders import load_star_systems, load_goods, load_equipment, load_ships
from world import StarSystem
from market import Good
from ship import Ship

# --- Helper for creating temp files with arbitrary content ---
def write_temp_json(data):
    fd, path = tempfile.mkstemp(suffix='.json')
    with os.fdopen(fd, 'w') as f:
        json.dump(data, f)
    return path

# --- Normal loader tests (from test_data_loading.py) ---
def test_load_star_systems():
    systems = load_star_systems(os.path.join('data', 'systems.json'))
    assert isinstance(systems, list)
    assert len(systems) > 0
    for s in systems:
        assert hasattr(s, 'name')
        assert hasattr(s, 'tech_level')
        assert hasattr(s, 'political_system')
        assert hasattr(s, 'x')
        assert hasattr(s, 'y')

def test_load_goods():
    goods = load_goods(os.path.join('data', 'goods.json'))
    assert isinstance(goods, list)
    assert len(goods) > 0
    for g in goods:
        assert hasattr(g, 'name')
        assert hasattr(g, 'base_price')
        assert hasattr(g, 'legality')
        assert hasattr(g, 'min_tech')
        assert hasattr(g, 'max_tech')

def test_load_equipment():
    equipment = load_equipment(os.path.join('data', 'equipment.json'))
    assert isinstance(equipment, list)
    assert len(equipment) > 0
    for item in equipment:
        assert 'name' in item
        assert 'type' in item
        assert 'tech_level' in item
        assert 'price' in item

def test_load_ships():
    ships = load_ships(os.path.join('data', 'ships.json'))
    assert isinstance(ships, list)
    assert len(ships) > 0
    for ship in ships:
        assert hasattr(ship, 'type')
        assert hasattr(ship, 'cargo_bays')
        assert hasattr(ship, 'hull')
        assert hasattr(ship, 'range')

# --- Robustness and edge case tests (from test_loaders_robust.py) ---
def test_load_star_systems_empty():
    path = write_temp_json([])
    systems = load_star_systems(path)
    assert systems == []
    os.remove(path)

def test_load_star_systems_malformed():
    path = write_temp_json([{"not_a_system": 1}])
    systems = load_star_systems(path)
    assert systems == [] or all(getattr(s, 'name', None) == '' for s in systems)
    os.remove(path)

def test_load_star_systems_missing_file():
    systems = load_star_systems('nonexistent_file.json')
    assert systems == []

def test_load_goods_empty():
    path = write_temp_json([])
    goods = load_goods(path)
    assert goods == []
    os.remove(path)

def test_load_goods_malformed():
    path = write_temp_json([{"foo": 123}])
    goods = load_goods(path)
    assert goods == []
    os.remove(path)

def test_load_goods_missing_file():
    goods = load_goods('nonexistent_file.json')
    assert goods == []

def test_load_equipment_empty():
    path = write_temp_json([])
    equipment = load_equipment(path)
    assert equipment == []
    os.remove(path)

def test_load_equipment_malformed():
    path = write_temp_json([{"foo": 123}])
    equipment = load_equipment(path)
    assert equipment == []
    os.remove(path)

def test_load_equipment_missing_file():
    equipment = load_equipment('nonexistent_file.json')
    assert equipment == []

def test_load_ships_empty():
    path = write_temp_json([])
    ships = load_ships(path)
    assert ships == []
    os.remove(path)

def test_load_ships_malformed():
    path = write_temp_json([{"foo": 123}])
    ships = load_ships(path)
    assert ships == []
    os.remove(path)

def test_load_ships_missing_file():
    ships = load_ships('nonexistent_file.json')
    assert ships == []

# --- Additional loader tests from loaders_tests.py ---
class TestStarSystemLoading:
    def setup_method(self):
        self.json_path = os.path.join(os.path.dirname(__file__), '../data', 'systems.json')

    def test_loads_all_systems(self):
        systems = load_star_systems(self.json_path)
        assert len(systems) > 0
        assert all(isinstance(s, StarSystem) for s in systems)

    def test_invalid_systems_rejected(self):
        invalids = [
            {},
            {'name': 'Test'},
            {'tech_level': 'Industrial'},
            {'political_system': 'Democracy'},
        ]
        for inv in invalids:
            sys = StarSystem.from_json_dict(inv)
            if 'name' not in inv:
                assert sys.name == ""
            if 'tech_level' not in inv:
                assert sys.tech_level == ""
            if 'political_system' not in inv:
                assert sys.political_system == ""

    def test_field_validity(self):
        valid_tech_levels = {
            "Pre-agricultural", "Agricultural", "Medieval", "Renaissance", "Early Industrial", "Industrial", "Post-industrial", "Hi-tech"
        }
        valid_political_systems = {
            "Anarchy", "Feudal State", "Confederacy", "Democracy", "Dictatorship", "Monarchy", "Technocracy", "Cybernetic State", "Capitalist State", "Corporate State"
        }
        systems = load_star_systems(self.json_path)
        for s in systems:
            assert isinstance(s.name, str) and len(s.name) > 0
            assert s.tech_level in valid_tech_levels
            assert s.political_system in valid_political_systems

    def test_required_fields(self):
        systems = load_star_systems(self.json_path)
        for s in systems:
            assert hasattr(s, 'name')
