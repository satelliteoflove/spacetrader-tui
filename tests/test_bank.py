import pytest
from player import Player
from bank import Bank

def test_take_loan_success():
    player = Player()
    player.loan_balance = 0
    player.credits = 1000
    ok, msg = Bank.take_loan(player, 5000)
    assert ok
    assert player.loan_balance == 5000
    assert player.credits == 6000
    assert "Loan of 5000 credits granted" in msg

def test_take_loan_over_limit():
    player = Player()
    player.loan_balance = 49000
    player.credits = 1000
    ok, msg = Bank.take_loan(player, 2000)
    assert not ok
    assert "Cannot borrow more than" in msg
    assert player.loan_balance == 49000
    assert player.credits == 1000

def test_repay_loan_partial():
    player = Player()
    player.loan_balance = 10000
    player.credits = 4000
    ok, msg = Bank.repay_loan(player, 3000)
    assert ok
    assert player.loan_balance == 7000
    assert player.credits == 1000
    assert "Repaid 3000 credits" in msg

def test_repay_loan_full():
    player = Player()
    player.loan_balance = 5000
    player.credits = 6000
    ok, msg = Bank.repay_loan(player, 5000)
    assert ok
    assert player.loan_balance == 0
    assert player.credits == 1000
    assert "Repaid 5000 credits" in msg

def test_repay_loan_overpay():
    player = Player()
    player.loan_balance = 3000
    player.credits = 5000
    ok, msg = Bank.repay_loan(player, 4000)
    assert ok
    assert player.loan_balance == 0
    assert player.credits == 2000
    assert "Repaid 3000 credits" in msg

def test_repay_loan_no_loan():
    player = Player()
    player.loan_balance = 0
    player.credits = 1000
    ok, msg = Bank.repay_loan(player, 1000)
    assert not ok
    assert "No loan to repay" in msg

def test_apply_interest():
    player = Player()
    player.loan_balance = 10000
    ok, msg = Bank.apply_interest(player)
    assert ok
    assert player.loan_balance == 10500  # 5% interest
    assert "Interest of 500 credits applied" in msg

def test_apply_interest_no_loan():
    player = Player()
    player.loan_balance = 0
    ok, msg = Bank.apply_interest(player)
    assert not ok
    assert "No loan to apply interest" in msg
