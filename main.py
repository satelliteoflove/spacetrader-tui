# Main entry point for Space Trader TUI
from game import GameState

def main():
    game = GameState()
    game.run()

if __name__ == "__main__":
    main()
