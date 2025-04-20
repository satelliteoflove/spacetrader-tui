from textual.screen import Screen
from textual.widgets import Static, Input, ListView, ListItem, Button
from textual.containers import Vertical, Container

class CharacterCreationScreen(Screen):
    def __init__(self):
        super().__init__()
        self.stage = 0  # 0: Name, 1: Skills, 2: Ship
        self.player_name = ""
        self.skills = {'pilot': 0, 'fighter': 0, 'trader': 0, 'engineer': 0}
        self.skill_points = 16
        self.selected_skill_idx = 0
        self.selected_ship = None
        self.available_ships = [
            {'type': 'Scout', 'hull': 8, 'max_hull': 8, 'fuel': 20, 'range': 5},
            {'type': 'Freighter', 'hull': 12, 'max_hull': 12, 'fuel': 15, 'range': 4},
            {'type': 'Fighter', 'hull': 10, 'max_hull': 10, 'fuel': 10, 'range': 3},
        ]

    def compose(self):
        if self.stage == 0:
            yield Static("=== CHARACTER CREATION ===\nEnter your name:", id="cc-title")
            yield Input(placeholder="Commander Name", id="name-input")
            yield Button("Continue", id="name-continue")
        elif self.stage == 1:
            yield Static(f"=== CHARACTER CREATION ===\nAllocate skill points ({self.skill_points} left):", id="cc-title")
            skill_items = [
                ListItem(Static(f"{skill.title()}: {val}"), id=skill)
                for skill, val in self.skills.items()
            ]
            yield ListView(*skill_items, id="skills-list")
            yield Button("Done", id="skills-done")
        elif self.stage == 2:
            yield Static("=== CHARACTER CREATION ===\nChoose your starting ship:", id="cc-title")
            ship_items = [
                ListItem(Static(f"{ship['type']} (Hull: {ship['hull']}, Fuel: {ship['fuel']}, Range: {ship['range']})"), id=str(i))
                for i, ship in enumerate(self.available_ships)
            ]
            yield ListView(*ship_items, id="ship-list")
            yield Button("Confirm", id="ship-confirm")

    async def on_button_pressed(self, event: Button.Pressed) -> None:
        if self.stage == 0 and event.button.id == "name-continue":
            name_input = self.query_one("#name-input", Input)
            self.player_name = name_input.value.strip()
            if self.player_name:
                self.stage = 1
                await self.app.pop_screen()
                await self.app.push_screen(CharacterCreationScreen())
            else:
                await self.app.push_screen(ErrorDialog("Name cannot be empty."))
        elif self.stage == 1 and event.button.id == "skills-done":
            if self.skill_points == 0:
                self.stage = 2
                await self.app.pop_screen()
                await self.app.push_screen(CharacterCreationScreen())
            else:
                await self.app.push_screen(ErrorDialog("Allocate all points before continuing."))
        elif self.stage == 2 and event.button.id == "ship-confirm":
            ship_list = self.query_one("#ship-list", ListView)
            idx = ship_list.index
            if 0 <= idx < len(self.available_ships):
                self.selected_ship = self.available_ships[idx]
                # Finalize creation and update game state
                # self.app.game_state.create_new_player(name=self.name, skills=self.skills, ship=self.selected_ship)
                await self.app.push_screen(SuccessDialog(f"Welcome, {self.player_name}!"))
                # await self.app.pop_screen() # To StatusScreen
            else:
                await self.app.push_screen(ErrorDialog("Invalid ship selection."))

    async def on_list_view_selected(self, event: ListView.Selected) -> None:
        if self.stage == 1:
            # Arrow key navigation for skills
            skill = event.item.id
            if self.skill_points > 0:
                self.skills[skill] += 1
                self.skill_points -= 1
                await self.app.pop_screen()
                await self.app.push_screen(CharacterCreationScreen())
        elif self.stage == 2:
            # Ship selection
            pass

    async def on_mount(self) -> None:
        if self.stage == 0:
            self.query_one("#name-input", Input).focus()
        elif self.stage == 1:
            self.query_one("#skills-list", ListView).focus()
        elif self.stage == 2:
            self.query_one("#ship-list", ListView).focus()

# ErrorDialog and SuccessDialog are placeholder screens for feedback.
class ErrorDialog(Screen):
    def __init__(self, message):
        super().__init__()
        self.message = message
    def compose(self):
        yield Static(f"Error: {self.message}")
        yield Button("OK", id="error-ok")
    async def on_button_pressed(self, event: Button.Pressed) -> None:
        await self.app.pop_screen()

class SuccessDialog(Screen):
    def __init__(self, message):
        super().__init__()
        self.message = message
    def compose(self):
        yield Static(self.message)
        yield Button("OK", id="success-ok")
    async def on_button_pressed(self, event: Button.Pressed) -> None:
        await self.app.pop_screen()


    def render(self):
        if self.stage == 0:
            print("\n=== CHARACTER CREATION ===\nEnter your name:")
        elif self.stage == 1:
            print("\n=== CHARACTER CREATION ===\nAllocate skill points ({} left):".format(self.skill_points))
            skills_list = list(self.skills.items())
            for idx, (skill, val) in enumerate(skills_list):
                if idx == self.selected_skill_idx:
                    # Highlight selected skill (e.g., with > and <)
                    print(f"> {skill.title()}: {val} <")
                else:
                    print(f"  {skill.title()}: {val}")
            print("Use UP/DOWN to select, LEFT/RIGHT to subtract/add points. Press ENTER when done.")
        elif self.stage == 2:
            print("\n=== CHARACTER CREATION ===\nChoose your starting ship:")
            for i, ship in enumerate(self.available_ships, 1):
                print(f"[{i}] {ship['type']} (Hull: {ship['hull']}, Fuel: {ship['fuel']}, Range: {ship['range']})")
            print("Enter the number of your choice.")

    def handle_input(self, user_input):
        if self.stage == 0:
            self.name = user_input.strip()
            if self.name:
                self.stage = 1
            else:
                print("Name cannot be empty.")
        elif self.stage == 1:
            # Expect arrow key input or ENTER
            key = user_input.lower()
            skills_list = list(self.skills.keys())
            if key in ("up", "k"):
                self.selected_skill_idx = (self.selected_skill_idx - 1) % len(skills_list)
            elif key in ("down", "j"):
                self.selected_skill_idx = (self.selected_skill_idx + 1) % len(skills_list)
            elif key in ("left", "h"):
                selected_skill = skills_list[self.selected_skill_idx]
                if self.skills[selected_skill] > 0:
                    self.skills[selected_skill] -= 1
                    self.skill_points += 1
            elif key in ("right", "l"):
                selected_skill = skills_list[self.selected_skill_idx]
                if self.skill_points > 0:
                    self.skills[selected_skill] += 1
                    self.skill_points -= 1
            elif key == "enter" or key == "\r" or key == "\n":
                if self.skill_points == 0:
                    self.stage = 2
                else:
                    print("Allocate all points before continuing.")
            else:
                print("Use arrow keys (or hjkl), and ENTER when done.")
        elif self.stage == 2:
            try:
                idx = int(user_input) - 1
                if 0 <= idx < len(self.available_ships):
                    self.selected_ship = self.available_ships[idx]
                    # Here, you would finalize creation and update game state
                    self.manager.game_state.create_new_player(
                        name=self.name,
                        skills=self.skills,
                        ship=self.selected_ship
                    )
                    print(f"Welcome, {self.player_name}!")
                    self.manager.switch_to(StatusScreen)
                else:
                    print("Invalid ship selection.")
            except Exception:
                print("Enter a valid number.")
