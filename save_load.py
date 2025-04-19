import json
from player import Player
from world import StarSystem


def save_game(filename, player, systems, meta=None):
    """
    Save the game state to a JSON file.
    Args:
        filename (str): Path to the save file.
        player (Player): The player object.
        systems (list[StarSystem]): List of all star systems.
        meta (dict): Optional meta info (turn, time, etc.)
            Should include: current_system (name), turn, time, random_seed, and other relevant fields.
    """
    meta = meta or {}
    # Try to auto-populate current_system if not set
    if 'current_system' not in meta and hasattr(player, 'current_system'):
        meta['current_system'] = getattr(player, 'current_system', None)
    if 'turn' not in meta:
        meta['turn'] = 1
    if 'time' not in meta:
        from datetime import datetime
        meta['time'] = datetime.now().isoformat()
    if 'random_seed' not in meta:
        meta['random_seed'] = 1234
    data = {
        'player': player.to_dict(),
        'systems': [s.to_dict() for s in systems],
        'meta': meta,
    }
    with open(filename, 'w', encoding='utf-8') as f:
        json.dump(data, f, indent=2)


def load_game(filename, goods_list=None):
    """
    Load the game state from a JSON file.
    Args:
        filename (str): Path to the save file.
        goods_list (list[Good]): List of Good objects, needed to reconstruct markets.
    Returns:
        (Player, list[StarSystem], dict): The player, systems, and meta info.
            meta will include: current_system, turn, time, random_seed, etc.
    """
    with open(filename, 'r', encoding='utf-8') as f:
        data = json.load(f)
    # Reconstruct player
    player = Player()
    player.from_dict(data['player'])
    # Reconstruct systems
    systems = []
    for sys_data in data['systems']:
        sys = StarSystem()
        sys.from_dict(sys_data)
        # Reconstruct market if goods_list provided
        if goods_list and sys.market:
            from market import Market
            mkt = Market(goods_list=goods_list, system=sys)
            mkt.from_dict(sys_data['market'])
            sys.market = mkt
        systems.append(sys)
    meta = data.get('meta', {})
    return player, systems, meta
