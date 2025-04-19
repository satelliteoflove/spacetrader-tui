from market import Market

import math

class StarSystem:
    """
    Represents a star system with its own market, attributes, and events.
    """
    def __init__(self, name="Sol", tech_level="Industrial", political_system="Democracy", resources=None, events=None, market=None, special=None, x=None, y=None, goods_list=None):
        self.name = name
        self.tech_level = tech_level
        self.political_system = political_system
        self.resources = resources if resources is not None else []
        self.events = events if events is not None else []
        # Always initialize market unless explicitly provided (for loading from save)
        if market:
            self.market = market
        elif goods_list:
            self.market = Market(goods_list=goods_list, system=self, events=self.events)
        else:
            self.market = None
        self.special = special
        self.x = x
        self.y = y

    def distance_to(self, other_system):
        """
        Return Euclidean distance to another StarSystem (in parsecs).
        """
        if self.x is None or self.y is None or other_system.x is None or other_system.y is None:
            raise ValueError("Both systems must have x and y coordinates.")
        return math.hypot(self.x - other_system.x, self.y - other_system.y)


    def to_dict(self):
        return {
            'name': self.name,
            'tech_level': self.tech_level,
            'political_system': self.political_system,
            'resources': self.resources,
            'events': self.events,
            'market': self.market.to_dict() if self.market else None,
        }

    def from_dict(self, data):
        self.name = data.get('name', "")
        self.tech_level = data.get('tech_level', "")
        self.political_system = data.get('political_system', "")
        # Accept both 'resource' (string) and 'resources' (list)
        if 'resources' in data:
            self.resources = data['resources']
        elif 'resource' in data:
            # Some JSONs use a single resource string
            self.resources = [data['resource']] if data['resource'] else []
        else:
            self.resources = []
        self.events = data.get('events', [])
        self.market = data.get('market', None)
        self.special = data.get('special', None)
        self.x = data.get('x', None)
        self.y = data.get('y', None)
        return self

    @classmethod
    def from_json_dict(cls, data):
        # Helper for direct instantiation from JSON dict
        name = data.get('name', "")
        tech_level = data.get('tech_level', "")
        political_system = data.get('political_system', "")
        if 'resources' in data:
            resources = data['resources']
        elif 'resource' in data:
            resources = [data['resource']] if data['resource'] else []
        else:
            resources = []
        events = data.get('events', [])
        market = data.get('market', None)
        special = data.get('special', None)
        x = data.get('x', None)
        y = data.get('y', None)
        return cls(name, tech_level, political_system, resources, events, market, special, x, y)
