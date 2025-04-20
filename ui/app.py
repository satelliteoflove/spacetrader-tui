# SpaceTraderApp: Textual-native TUI entry point
# References: https://textual.textualize.io/guide/app/ and https://textual.textualize.io/guide/screens/

from textual.app import App, ComposeResult
from textual.widgets import Header, Footer, Static
from textual.screen import Screen
from textual.containers import Container
from ui.screens import MainMenuScreen

class SpaceTraderApp(App):
    CSS_PATH = "app.tcss"  # Optional: for consistent style

    def compose(self) -> ComposeResult:
        yield Header()
        yield Footer()
        yield Container(id="main-container")

    def on_mount(self) -> None:
        self.push_screen(MainMenuScreen())

if __name__ == "__main__":
    SpaceTraderApp().run()
