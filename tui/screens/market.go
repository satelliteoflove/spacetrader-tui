package screens

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/the4ofus/spacetrader-tui/internal/game"
	"github.com/the4ofus/spacetrader-tui/internal/market"
)

type marketMode int

const (
	modeBrowse marketMode = iota
	modeBuyQty
	modeSellQty
)

type MarketScreen struct {
	gs        *game.GameState
	cursor    int
	mode      marketMode
	qtyInput  textinput.Model
	message   string
	goods     []int
	avgPrices [game.NumGoods]int
}

func NewMarketScreen(gs *game.GameState) *MarketScreen {
	ti := textinput.New()
	ti.Placeholder = "qty"
	ti.CharLimit = 4

	sysState := gs.Systems[gs.CurrentSystemID]
	var goods []int
	for i := 0; i < game.NumGoods; i++ {
		if sysState.Prices[i] > 0 || gs.Player.Cargo[i] > 0 {
			goods = append(goods, i)
		}
	}

	var avgPrices [game.NumGoods]int
	for i, good := range gs.Data.Goods {
		avgPrices[i] = market.AveragePrice(good, gs.Data.Systems)
	}

	return &MarketScreen{
		gs:        gs,
		qtyInput:  ti,
		goods:     goods,
		avgPrices: avgPrices,
	}
}

func (s *MarketScreen) Init() tea.Cmd { return nil }

func (s *MarketScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch s.mode {
	case modeBrowse:
		return s.updateBrowse(msg)
	case modeBuyQty, modeSellQty:
		return s.updateQty(msg)
	}
	return s, nil
}

func (s *MarketScreen) updateBrowse(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, Keys.Up):
			s.cursor = wrapCursor(s.cursor, -1, len(s.goods))
		case key.Matches(msg, Keys.Down):
			s.cursor = wrapCursor(s.cursor, 1, len(s.goods))
		case msg.String() == "b":
			if len(s.goods) > 0 {
				goodIdx := s.goods[s.cursor]
				if s.gs.Systems[s.gs.CurrentSystemID].Prices[goodIdx] > 0 {
					s.mode = modeBuyQty
					s.qtyInput.Reset()
					s.qtyInput.Focus()
					s.message = ""
					return s, textinput.Blink
				}
			}
		case msg.String() == "s":
			if len(s.goods) > 0 {
				goodIdx := s.goods[s.cursor]
				if s.gs.Player.Cargo[goodIdx] > 0 {
					s.mode = modeSellQty
					s.qtyInput.Reset()
					s.qtyInput.Focus()
					s.message = ""
					return s, textinput.Blink
				}
			}
		case key.Matches(msg, Keys.Back):
			return s, func() tea.Msg { return NavigateMsg{Screen: ScreenSystem} }
		}
	}
	return s, nil
}

func (s *MarketScreen) updateQty(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if key.Matches(msg, Keys.Back) {
			s.mode = modeBrowse
			s.message = ""
			return s, nil
		}
		if key.Matches(msg, Keys.Enter) {
			qty, err := strconv.Atoi(strings.TrimSpace(s.qtyInput.Value()))
			if err != nil || qty <= 0 {
				s.message = "Invalid quantity."
				s.mode = modeBrowse
				return s, nil
			}

			goodIdx := s.goods[s.cursor]
			var result market.TransactionResult
			if s.mode == modeBuyQty {
				result = market.Buy(s.gs, goodIdx, qty)
			} else {
				result = market.Sell(s.gs, goodIdx, qty)
			}
			s.message = result.Message
			s.mode = modeBrowse
			return s, nil
		}
	}
	var cmd tea.Cmd
	s.qtyInput, cmd = s.qtyInput.Update(msg)
	return s, cmd
}

func (s *MarketScreen) View() string {
	var b strings.Builder

	dp := &game.GameDataProvider{Data: s.gs.Data}
	freeCargo := s.gs.Player.FreeCargo(dp)
	sysState := s.gs.Systems[s.gs.CurrentSystemID]
	sysName := s.gs.Data.Systems[s.gs.CurrentSystemID].Name

	b.WriteString(HeaderStyle.Render(fmt.Sprintf("  %s - LOCAL MARKET  ", sysName)) + "\n")
	b.WriteString(fmt.Sprintf("  Credits: %d  |  Cargo: %d/%d (%d free)\n\n",
		s.gs.Player.Credits, s.gs.Player.TotalCargo(),
		s.gs.Player.CargoCapacity(dp), freeCargo))

	//  header row
	b.WriteString(DimStyle.Render(fmt.Sprintf(
		"     %-12s %6s  %4s  %6s  %s",
		"GOOD", "PRICE", "HELD", "AVG", "TREND")) + "\n")
	b.WriteString("  " + strings.Repeat("-", 50) + "\n")

	for i, goodIdx := range s.goods {
		good := s.gs.Data.Goods[goodIdx]
		price := sysState.Prices[goodIdx]
		owned := s.gs.Player.Cargo[goodIdx]
		avg := s.avgPrices[goodIdx]

		priceStr := fmt.Sprintf("%d", price)
		avgStr := fmt.Sprintf("%d", avg)
		if price < 0 {
			priceStr = "--"
			avgStr = "--"
		}

		trend := ""
		if price > 0 {
			hint := market.PriceVsAverage(price, avg)
			switch hint {
			case "very cheap":
				trend = SuccessStyle.Render("<<")
			case "cheap":
				trend = SuccessStyle.Render("<")
			case "very expensive":
				trend = DangerStyle.Render(">>")
			case "expensive":
				trend = DangerStyle.Render(">")
			default:
				trend = DimStyle.Render("=")
			}
		}

		name := good.Name
		illegal := ""
		if !good.Legal {
			illegal = DangerStyle.Render("!")
		}

		heldStr := fmt.Sprintf("%d", owned)
		if owned == 0 {
			heldStr = DimStyle.Render("-")
		}

		row := fmt.Sprintf("%-12s%s %6s  %4s  %6s   %s",
			name, illegal, priceStr, heldStr, avgStr, trend)

		if i == s.cursor {
			b.WriteString(SelectedStyle.Render("> ") + row + "\n")
		} else {
			b.WriteString("  " + row + "\n")
		}
	}

	b.WriteString("\n")

	if s.mode == modeBuyQty || s.mode == modeSellQty {
		goodIdx := s.goods[s.cursor]
		price := sysState.Prices[goodIdx]
		goodName := s.gs.Data.Goods[goodIdx].Name

		if s.mode == modeBuyQty {
			b.WriteString(fmt.Sprintf("  Buy %s @ %d cr: ", goodName, price))
		} else {
			b.WriteString(fmt.Sprintf("  Sell %s @ %d cr: ", goodName, price))
		}
		b.WriteString(s.qtyInput.View() + "\n")

		qtyStr := strings.TrimSpace(s.qtyInput.Value())
		if qty, err := strconv.Atoi(qtyStr); err == nil && qty > 0 {
			total := price * qty
			if s.mode == modeBuyQty {
				remaining := s.gs.Player.Credits - total
				if remaining < 0 {
					b.WriteString(DangerStyle.Render(
						fmt.Sprintf("  Total: %d cr  |  Short %d cr", total, -remaining)) + "\n")
				} else if qty > freeCargo {
					b.WriteString(DangerStyle.Render(
						fmt.Sprintf("  Total: %d cr  |  Need %d bays, have %d", total, qty, freeCargo)) + "\n")
				} else {
					b.WriteString(fmt.Sprintf("  Total: %d cr  |  After: %d cr\n", total, remaining))
				}
			} else {
				held := s.gs.Player.Cargo[goodIdx]
				if qty > held {
					b.WriteString(DangerStyle.Render(
						fmt.Sprintf("  Revenue: %d cr  |  Only have %d", total, held)) + "\n")
				} else {
					after := s.gs.Player.Credits + total
					b.WriteString(fmt.Sprintf("  Revenue: %d cr  |  After: %d cr\n", total, after))
				}
			}
		}
	}

	if s.message != "" {
		b.WriteString("  " + s.message + "\n")
	}

	b.WriteString("\n" + DimStyle.Render("  b buy  s sell  << cheap  >> pricey  ! illegal  esc back"))
	return b.String()
}
