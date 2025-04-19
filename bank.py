class Bank:
    """
    Handles loans, repayments, and interest for the player.
    """
    MAX_LOAN = 50000
    INTEREST_RATE = 0.05  # 5% per period (e.g., per travel)

    @staticmethod
    def take_loan(player, amount):
        if amount <= 0:
            return False, "Loan amount must be positive."
        if player.loan_balance + amount > Bank.MAX_LOAN:
            return False, f"Cannot borrow more than {Bank.MAX_LOAN} credits."
        player.loan_balance += amount
        player.credits += amount
        return True, f"Loan of {amount} credits granted. Total loan: {player.loan_balance}."

    @staticmethod
    def repay_loan(player, amount):
        if amount <= 0:
            return False, "Repayment amount must be positive."
        if player.loan_balance == 0:
            return False, "No loan to repay."
        pay = min(amount, player.loan_balance, player.credits)
        if pay == 0:
            return False, "Not enough credits to repay loan."
        player.loan_balance -= pay
        player.credits -= pay
        return True, f"Repaid {pay} credits. Remaining loan: {player.loan_balance}."

    @staticmethod
    def apply_interest(player):
        if player.loan_balance > 0:
            interest = int(player.loan_balance * Bank.INTEREST_RATE)
            player.loan_balance += interest
            return True, f"Interest of {interest} credits applied. Total loan: {player.loan_balance}."
        return False, "No loan to apply interest to."
