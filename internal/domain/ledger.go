package domain

import (
	"errors"
	"time"
)

var (
	ErrInvalidLedgerEntry = errors.New("invalid ledger entry")
)

type LedgerDirection string

const (
	LedgerDebit  LedgerDirection = "DEBIT"
	LedgerCredit LedgerDirection = "CREDIT"
)

type WalletLedgerEntry struct {
	ID            string
	WalletID      string
	TransactionID string
	Direction     LedgerDirection
	Amount        Money
	BalanceBefore Money
	BalanceAfter  Money
	CreatedAt     time.Time
}

func NewWalletLedgerEntry(
	id string,
	walletID string,
	transactionID string,
	direction LedgerDirection,
	amount Money,
	balanceBefore Money,
	balanceAfter Money,
	createdAt time.Time,
) (WalletLedgerEntry, error) {
	if id == "" {
		return WalletLedgerEntry{}, ErrInvalidLedgerEntry
	}

	if walletID == "" {
		return WalletLedgerEntry{}, ErrInvalidLedgerEntry
	}

	if transactionID == "" {
		return WalletLedgerEntry{}, ErrInvalidLedgerEntry
	}

	if createdAt.IsZero() {
		return WalletLedgerEntry{}, ErrInvalidLedgerEntry
	}

	switch direction {
	case LedgerDebit, LedgerCredit:
		// direção válida
	default:
		return WalletLedgerEntry{}, ErrInvalidLedgerEntry
	}

	if amount.Amount <= 0 {
		return WalletLedgerEntry{}, ErrInvalidLedgerEntry
	}

	if amount.Currency != balanceBefore.Currency ||
		amount.Currency != balanceAfter.Currency {
		return WalletLedgerEntry{}, ErrCurrencyMismatch
	}

	var expectedBalance Money
	var err error

	switch direction {
	case LedgerDebit:
		expectedBalance, err = balanceBefore.Subtract(amount)
		if err != nil {
			return WalletLedgerEntry{}, err
		}

	case LedgerCredit:
		expectedBalance, err = balanceBefore.Add(amount)
		if err != nil {
			return WalletLedgerEntry{}, err
		}
	}

	if expectedBalance.Amount != balanceAfter.Amount ||
		expectedBalance.Currency != balanceAfter.Currency {
		return WalletLedgerEntry{}, ErrInvalidLedgerEntry
	}

	if balanceBefore.Amount < 0 || balanceAfter.Amount < 0 {
		return WalletLedgerEntry{}, ErrInvalidLedgerEntry
	}

	return WalletLedgerEntry{
		ID:            id,
		WalletID:      walletID,
		TransactionID: transactionID,
		Direction:     direction,
		Amount:        amount,
		BalanceBefore: balanceBefore,
		BalanceAfter:  balanceAfter,
		CreatedAt:     createdAt,
	}, nil
}
