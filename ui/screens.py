# ui/screens.py
# Text-based TUI screen classes for Space Trader

# ScreenManager is no longer needed; navigation is handled by Textual's App and Screen system.
# Kept for legacy reference, but not used in Textual-native refactor.

from ui.character_creation import CharacterCreationScreen

from textual.screen import Screen
from textual.widgets import ListView, ListItem, Static, Input, Button
from textual.containers import Container, Vertical
from textual import events
from ui.widgets import PlayerStatusWidget

class MainMenuScreen(Screen):
    def compose(self):
        yield Static("SPACE TRADER - TERMINAL EDITION", id="main-title")
        yield ListView(
            ListItem(Static("Start New Game"), id="start"),
            ListItem(Static("Load Game"), id="load"),
            ListItem(Static("Quit"), id="quit"),
            id="main-menu-list"
        )
        # Placeholder for PlayerStatusWidget
        # yield PlayerStatusWidget()

    async def on_list_view_selected(self, event: ListView.Selected) -> None:
        selection = event.item.id
        if selection == "start":
            from ui.character_creation import CharacterCreationScreen
            await self.app.push_screen(CharacterCreationScreen())
        elif selection == "load":
            # Show input dialog for filename (Textual Input)
            await self.app.push_screen(LoadGameScreen())
        elif selection == "quit":
            await self.app.action_quit()

    async def on_mount(self) -> None:
        menu = self.query_one("#main-menu-list", ListView)
        menu.focus()

class BaseScreen:
    def render_player_status(self):
        p = self.player
        print(f"""
========================================
       COMMANDER STATUS
========================================
Name: {p.name}         Credits: {p.credits}
Ship: {p.ship.type}          Hull: {p.ship.hull}/{p.ship.max_hull}
Skills: Pilot({p.skills.get('pilot', 0)})  Fighter({p.skills.get('fighter', 0)})  Trader({p.skills.get('trader', 0)})  Engineer({p.skills.get('engineer', 0)})
Police Record: {getattr(p, 'police_record', 0)}    Reputation: {getattr(p, 'reputation', 0)}
""")

from textual.screen import Screen
from textual.widgets import ListView, ListItem, Static, Input, Button
from textual.containers import Container, Vertical

class StatusScreen(Screen):
    def __init__(self, player=None):
        super().__init__()
        self.player = player

    def compose(self):
        if self.player:
            yield PlayerStatusWidget(self.player)
        else:
            yield Static("Commander status unavailable.", id="player-status")
        yield ListView(
            ListItem(Static("View Market"), id="market"),
            ListItem(Static("Travel"), id="travel"),
            ListItem(Static("Shipyard"), id="shipyard"),
            ListItem(Static("Bank"), id="bank"),
            ListItem(Static("Save Game"), id="save"),
            ListItem(Static("Main Menu"), id="menu"),
            id="status-menu-list"
        )

    async def on_list_view_selected(self, event: ListView.Selected) -> None:
        selection = event.item.id
        if selection == "market":
            from ui.screens import MarketScreen
            await self.app.push_screen(MarketScreen(self.player))
        elif selection == "travel":
            from ui.screens import TravelScreen
            await self.app.push_screen(TravelScreen(self.player))
        elif selection == "shipyard":
            from ui.screens import ShipyardScreen
            await self.app.push_screen(ShipyardScreen(self.player))
        elif selection == "bank":
            from ui.screens import BankScreen
            await self.app.push_screen(BankScreen(self.player))
        elif selection == "save":
            await self.app.push_screen(SaveGameScreen(self.player))
        elif selection == "menu":
            from ui.screens import MainMenuScreen
            await self.app.push_screen(MainMenuScreen())

    async def on_mount(self) -> None:
        menu = self.query_one("#status-menu-list", ListView)
        await menu.focus()

class MarketScreen(Screen):
    def __init__(self, player=None):
        super().__init__()
        self.player = player
        self.market = getattr(getattr(player, 'system', None), 'market', None)

    def compose(self):
        if self.player:
            yield PlayerStatusWidget(self.player)
        yield Static("LOCAL MARKET", id="market-title")
        if not self.market:
            yield Static("No market data available.", id="market-empty")
        else:
            items = []
            for good in self.market.goods.values():
                owned = self.player.inventory.get(good.name, 0)
                price = self.market.prices.get(good.name, '?')
                label = f"{good.name:<13} {price:<8} {owned:<7} [B]   [S]{' *' if not good.legality else ''}"
                items.append(ListItem(Static(label), id=good.name))
            yield ListView(*items, id="market-list")
        yield Input(placeholder="B Water 3 or S Water 3", id="trade-input")
        yield Button("Back", id="market-back")

    async def on_button_pressed(self, event: Button.Pressed) -> None:
        if event.button.id == "market-back":
            from ui.screens import StatusScreen
            await self.app.push_screen(StatusScreen(self.player))

    async def on_input_submitted(self, event: Input.Submitted) -> None:
        tokens = event.value.strip().split()
        if len(tokens) == 3 and tokens[0].upper() in ('B', 'S'):
            action, good_name, qty = tokens[0].upper(), tokens[1], tokens[2]
            try:
                qty = int(qty)
                if action == 'B':
                    success, msg = self.player.buy_good(self.market, good_name, qty)
                else:
                    success, msg = self.player.sell_good(self.market, good_name, qty)
                await self.app.push_screen(SuccessDialog(msg))
            except Exception as e:
                await self.app.push_screen(ErrorDialog(f"Error: {e}"))
        else:
            await self.app.push_screen(ErrorDialog("Invalid input."))

    async def on_mount(self) -> None:
        await self.query_one("#trade-input", Input).focus()

class TravelScreen(Screen):
    def __init__(self, player=None):
        super().__init__()
        self.player = player
        self.system = getattr(player, 'system', None)
        self.reachable = getattr(getattr(player, 'game_state', None), 'get_reachable_systems', lambda: [])()

    def compose(self):
        if self.player:
            yield PlayerStatusWidget(self.player)
        yield Static("GALAXY MAP", id="travel-title")
        yield Static(f"Current System: {self.system.name if self.system else '?'}", id="travel-system")
        yield Static(f"Fuel: {self.player.ship.fuel}/{self.player.ship.range}", id="travel-fuel")
        items = []
        for i, (sys, dist) in enumerate(self.reachable, 1):
            label = f"[{i}] {sys.name:<15} ({dist:.1f} parsecs)"
            items.append(ListItem(Static(label), id=str(i-1)))
        yield ListView(*items, id="travel-list")
        yield Button("Back", id="travel-back")

    async def on_button_pressed(self, event: Button.Pressed) -> None:
        if event.button.id == "travel-back":
            from ui.screens import StatusScreen
            await self.app.push_screen(StatusScreen(self.player))

    async def on_list_view_selected(self, event: ListView.Selected) -> None:
        idx = int(event.item.id)
        if 0 <= idx < len(self.reachable):
            dest, dist = self.reachable[idx]
            success, msg = self.player.game_state.travel_to_system(dest)
            if not success:
                await self.app.push_screen(ErrorDialog(msg))
                return
            # Check for encounter
            from encounter import Encounter
            encounter = Encounter.random_encounter(self.player, dest)
            if encounter:
                from ui.screens import EncounterScreen
                await self.app.push_screen(EncounterScreen(self.player, encounter))
            else:
                from ui.screens import StatusScreen
                await self.app.push_screen(StatusScreen(self.player))

    async def on_mount(self) -> None:
        await self.query_one("#travel-list", ListView).focus()


class EncounterScreen(Screen):
    def __init__(self, player=None, encounter=None):
        super().__init__()
        self.player = player
        self.encounter = encounter
        self.system = getattr(player, 'system', None)

    def compose(self):
        if self.player:
            yield PlayerStatusWidget(self.player)
        yield Static(f"ENCOUNTER: {self.encounter.type.upper()}", id="encounter-title")
        options = []
        if self.encounter.type == 'police':
            yield Static("Police are hailing your ship for inspection.")
            options = [
                ("Comply", "comply"),
                ("Attempt Bribe", "bribe"),
                ("Attempt to Flee", "flee")
            ]
        elif self.encounter.type == 'pirate':
            yield Static("Pirates block your path!")
            options = [
                ("Fight", "attack"),
                ("Attempt to Flee", "flee"),
                ("Surrender", "surrender")
            ]
        elif self.encounter.type == 'trader':
            yield Static("A trader offers to deal.")
            options = [
                ("Trade", "trade"),
                ("Decline", "decline")
            ]
        else:
            yield Static("Unknown encounter.")
        yield ListView(
            *(ListItem(Static(label), id=action) for label, action in options),
            id="encounter-list"
        )

    async def on_list_view_selected(self, event: ListView.Selected) -> None:
        action = event.item.id
        msg = self.encounter.resolve(self.player, self.system, action)
        await self.app.push_screen(SuccessDialog(msg))
        from ui.screens import StatusScreen
        await self.app.push_screen(StatusScreen(self.player))

    async def on_mount(self) -> None:
        await self.query_one("#encounter-list", ListView).focus()


class ShipyardScreen(Screen):
    def __init__(self, player=None):
        super().__init__()
        self.player = player
        # Placeholder: could load available ships from game_state

    def compose(self):
        if self.player:
            yield PlayerStatusWidget(self.player)
        yield Static("SHIPYARD", id="shipyard-title")
        yield ListView(
            ListItem(Static("View/Upgrade Ship"), id="view"),
            ListItem(Static("Buy New Ship (not implemented)"), id="buy"),
            ListItem(Static("Back"), id="back"),
            id="shipyard-list"
        )

    async def on_list_view_selected(self, event: ListView.Selected) -> None:
        if event.item.id == "view":
            await self.app.push_screen(SuccessDialog(f"Ship: {self.player.ship.type}, Hull: {self.player.ship.hull}/{self.player.ship.max_hull}\nEquipment: {self.player.ship.equipment}"))
        elif event.item.id == "buy":
            await self.app.push_screen(ErrorDialog("Buying new ships is not implemented yet."))
        elif event.item.id == "back":
            from ui.screens import StatusScreen
            await self.app.push_screen(StatusScreen(self.player))

    async def on_mount(self) -> None:
        await self.query_one("#shipyard-list", ListView).focus()


class BankScreen(Screen):
    def __init__(self, player=None):
        super().__init__()
        self.player = player

    def compose(self):
        if self.player:
            yield PlayerStatusWidget(self.player)
        yield Static("BANK", id="bank-title")
        yield Static(f"Credits: {self.player.credits}", id="bank-credits")
        yield Static(f"Loan Balance: {getattr(self.player, 'loan_balance', 0)}", id="bank-loan")
        yield ListView(
            ListItem(Static("Take Loan (not implemented)"), id="loan"),
            ListItem(Static("Repay Loan (not implemented)"), id="repay"),
            ListItem(Static("Back"), id="back"),
            id="bank-list"
        )

    async def on_list_view_selected(self, event: ListView.Selected) -> None:
        if event.item.id == "loan":
            await self.app.push_screen(ErrorDialog("Loans not implemented yet."))
        elif event.item.id == "repay":
            await self.app.push_screen(ErrorDialog("Repaying loans not implemented yet."))
        elif event.item.id == "back":
            from ui.screens import StatusScreen
            await self.app.push_screen(StatusScreen(self.player))

    async def on_mount(self) -> None:
        await self.query_one("#bank-list", ListView).focus()


class PersonnelScreen(Screen):
    def __init__(self, player=None):
        super().__init__()
        self.player = player

    def compose(self):
        if self.player:
            yield PlayerStatusWidget(self.player)
        yield Static("CREW MANAGEMENT", id="personnel-title")
        yield ListView(
            ListItem(Static("View Crew (not implemented)"), id="view"),
            ListItem(Static("Hire Mercenary (not implemented)"), id="hire"),
            ListItem(Static("Fire Mercenary (not implemented)"), id="fire"),
            ListItem(Static("Back"), id="back"),
            id="personnel-list"
        )

    async def on_list_view_selected(self, event: ListView.Selected) -> None:
        if event.item.id == "back":
            from ui.screens import StatusScreen
            await self.app.push_screen(StatusScreen(self.player))
        else:
            await self.app.push_screen(ErrorDialog("Not implemented yet."))

    async def on_mount(self) -> None:
        await self.query_one("#personnel-list", ListView).focus()