import json
from world import StarSystem
from market import Good
from ship import Ship

def load_star_systems(json_path):
    """Load all star systems from a JSON file and return a list of StarSystem objects."""
    import logging
    try:
        with open(json_path, 'r') as f:
            systems_data = json.load(f)
        systems = []
        for idx, sd in enumerate(systems_data):
            if not isinstance(sd, dict):
                logging.error(f"Star system at index {idx} is not a dict: {sd}")
                continue
            required = ["name", "tech_level", "political_system", "x", "y"]
            for field in required:
                if field not in sd:
                    logging.error(f"Missing required field '{field}' in star system at index {idx}: {sd}")
                    break
            try:
                systems.append(StarSystem.from_json_dict(sd))
            except Exception as e:
                logging.error(f"Failed to instantiate StarSystem at index {idx}: {e}")
        return systems
    except Exception as e:
        logging.error(f"Failed to load star systems from {json_path}: {e}")
        return []

def load_goods(json_path):
    """Load all goods from a JSON file and return a list of Good objects."""
    import logging
    try:
        with open(json_path, 'r') as f:
            goods_data = json.load(f)
        goods = []
        for idx, gd in enumerate(goods_data):
            if not isinstance(gd, dict):
                logging.error(f"Good at index {idx} is not a dict: {gd}")
                continue
            required = ["name", "base_price", "legality", "min_tech", "max_tech"]
            for field in required:
                if field not in gd:
                    logging.error(f"Missing required field '{field}' in good at index {idx}: {gd}")
                    break
            try:
                goods.append(Good(**gd))
            except Exception as e:
                logging.error(f"Failed to instantiate Good at index {idx}: {e}")
        return goods
    except Exception as e:
        logging.error(f"Failed to load goods from {json_path}: {e}")
        return []

def load_equipment(json_path):
    """Load all equipment from a JSON file and return a list of equipment dicts."""
    import logging
    try:
        with open(json_path, 'r') as f:
            equipment_data = json.load(f)
        equipment = []
        for idx, eq in enumerate(equipment_data):
            if not isinstance(eq, dict):
                logging.error(f"Equipment at index {idx} is not a dict: {eq}")
                continue
            required = ["name", "type", "tech_level", "price"]
            skip = False
            for field in required:
                if field not in eq:
                    logging.error(f"Missing required field '{field}' in equipment at index {idx}: {eq}")
                    skip = True
                    break
            if skip:
                continue
            equipment.append(eq)
        return equipment
    except Exception as e:
        logging.error(f"Failed to load equipment from {json_path}: {e}")
        return []

def load_ships(json_path):
    """Load all ships from a JSON file and return a list of Ship objects."""
    import logging
    try:
        with open(json_path, 'r') as f:
            ships_data = json.load(f)
        ships = []
        for idx, sd in enumerate(ships_data):
            if not isinstance(sd, dict):
                logging.error(f"Ship at index {idx} is not a dict: {sd}")
                continue
            required = ["type", "cargo_bays", "hull", "range"]
            skip = False
            for field in required:
                if field not in sd:
                    logging.error(f"Missing required field '{field}' in ship at index {idx}: {sd}")
                    skip = True
                    break
            if skip:
                continue
            try:
                ships.append(Ship().from_dict(sd))
            except Exception as e:
                logging.error(f"Failed to instantiate Ship at index {idx}: {e}")
        return ships
    except Exception as e:
        logging.error(f"Failed to load ships from {json_path}: {e}")
        return []
