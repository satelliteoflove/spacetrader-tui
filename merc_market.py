from mercenary import Mercenary
import random

# Sample pool of mercenaries (in real game, could be loaded from data)
SAMPLE_MERCS = [
    Mercenary("Axel Steel", "fighter", {"fighter": 4}, 1500),
    Mercenary("Luna Wise", "pilot", {"pilot": 3}, 1200),
    Mercenary("Mira Bolt", "engineer", {"engineer": 5}, 2000),
    Mercenary("Dex Trader", "trader", {"trader": 4}, 1600),
    Mercenary("Jax Quick", "pilot", {"pilot": 2, "fighter": 1}, 1000),
]

def get_available_mercenaries(max_count=3):
    """Return a random sample of available mercenaries for hire."""
    return random.sample(SAMPLE_MERCS, k=min(max_count, len(SAMPLE_MERCS)))
