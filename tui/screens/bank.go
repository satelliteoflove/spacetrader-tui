package screens

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/the4ofus/spacetrader-tui/internal/economy"
	"github.com/the4ofus/spacetrader-tui/internal/game"
)

type bankMode int

const (
	bankMenu bankMode = iota
	bankBorrow
	bankRepay
)

type BankScreen struct {
	gs       *game.GameState
	mode     bankMode
	cursor   int
	input    textinput.Model
	message  string
}

func NewBankScreen(gs *game.GameState) *BankScreen {
	ti := textinput.New()
	ti.Placeholder = "amount"
	ti.CharLimit = 8
	return &BankScreen{gs: gs, input: ti}
}

func (s *BankScreen) Init() tea.Cmd { return nil }

func (s *BankScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if s.mode == bankBorrow || s.mode == bankRepay {
		return s.updateInput(msg)
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, Keys.Up):
			s.cursor = wrapCursor(s.cursor, -1, 2)
		case key.Matches(msg, Keys.Down):
			s.cursor = wrapCursor(s.cursor, 1, 2)
		case key.Matches(msg, Keys.Enter):
			if s.cursor == 0 {
				s.mode = bankBorrow
				s.input.Reset()
				s.input.Focus()
				return s, textinput.Blink
			} else {
				s.mode = bankRepay
				s.input.Reset()
				s.input.Focus()
				return s, textinput.Blink
			}
		case key.Matches(msg, Keys.Back):
			return s, func() tea.Msg { return NavigateMsg{Screen: ScreenSystem} }
		}
	}
	return s, nil
}

func (s *BankScreen) updateInput(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if key.Matches(msg, Keys.Back) {
			s.mode = bankMenu
			return s, nil
		}
		if key.Matches(msg, Keys.Enter) {
			amount, err := strconv.Atoi(strings.TrimSpace(s.input.Value()))
			if err != nil || amount <= 0 {
				s.message = "Invalid amount."
				s.mode = bankMenu
				return s, nil
			}

			if s.mode == bankBorrow {
				result := economy.TakeLoan(s.gs, amount)
				s.message = result.Message
			} else {
				result := economy.RepayLoan(s.gs, amount)
				s.message = result.Message
			}
			s.mode = bankMenu
			return s, nil
		}
	}
	var cmd tea.Cmd
	s.input, cmd = s.input.Update(msg)
	return s, cmd
}

func (s *BankScreen) View() string {
	var b strings.Builder

	b.WriteString(HeaderStyle.Render("  BANK  ") + "\n")
	b.WriteString(fmt.Sprintf("  Credits: %d\n", s.gs.Player.Credits))
	b.WriteString(fmt.Sprintf("  Loan balance: %d\n", s.gs.Player.LoanBalance))
	b.WriteString(fmt.Sprintf("  Max loan: %d\n", economy.MaxLoanAmount(s.gs)))
	b.WriteString(fmt.Sprintf("  Interest rate: 10%% per warp\n"))

	if s.gs.Player.LoanBalance > 0 {
		bal := s.gs.Player.LoanBalance
		b.WriteString(DimStyle.Render("  Projected if unpaid:") + "\n")
		b.WriteString(DimStyle.Render(fmt.Sprintf("    After 1 warp:   %d cr", economy.ProjectLoan(bal, 1))) + "\n")
		b.WriteString(DimStyle.Render(fmt.Sprintf("    After 5 warps:  %d cr", economy.ProjectLoan(bal, 5))) + "\n")
		b.WriteString(DimStyle.Render(fmt.Sprintf("    After 10 warps: %d cr", economy.ProjectLoan(bal, 10))) + "\n")
	}
	b.WriteString("\n")

	items := []string{"Borrow credits", "Repay loan"}
	RenderMenuItems(&b, items, s.cursor)

	if s.mode == bankBorrow || s.mode == bankRepay {
		label := "Borrow"
		if s.mode == bankRepay {
			label = "Repay"
		}
		b.WriteString(fmt.Sprintf("\n  %s amount: %s", label, s.input.View()))
	}

	if s.message != "" {
		b.WriteString("\n  " + s.message)
	}

	b.WriteString("\n\n" + DimStyle.Render("  j/k navigate, enter select, esc back, ? help"))
	return b.String()
}

func (s *BankScreen) HelpTitle() string { return "Bank" }

func (s *BankScreen) HelpGroups() []KeyGroup {
	return []KeyGroup{
		{
			Title: "Actions",
			Bindings: []KeyBinding{
				{Keys: "j/k or arrows", Desc: "Move cursor"},
				{Keys: "enter", Desc: "Borrow or repay"},
			},
		},
		{
			Title: "Notes",
			Bindings: []KeyBinding{
				{Keys: "", Desc: "Interest is 10% per warp, compounded."},
				{Keys: "", Desc: "Projected debt is shown when a loan is active."},
			},
		},
	}
}
