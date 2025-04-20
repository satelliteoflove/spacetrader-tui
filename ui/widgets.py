# ui/widgets.py
# Placeholder for reusable UI widgets

from textual.widgets import Static

class PlayerStatusWidget(Static):
    def __init__(self, player, **kwargs):
        self.player = player
        super().__init__(self._render_status(), **kwargs)

    def _render_status(self):
        p = self.player
        return (
            f"Commander: {p.name}\n"
            f"Credits: {p.credits}\n"
            f"Ship: {p.ship.type} (Hull: {p.ship.hull}/{p.ship.max_hull})\n"
            f"Skills: Pilot({p.skills.get('pilot', 0)}) "
            f"Fighter({p.skills.get('fighter', 0)}) "
            f"Trader({p.skills.get('trader', 0)}) "
            f"Engineer({p.skills.get('engineer', 0)})\n"
            f"Police Record: {getattr(p, 'police_record', 0)} "
            f"Reputation: {getattr(p, 'reputation', 0)}"
        )

    def update_player(self, player):
        self.player = player
        self.update(self._render_status())
