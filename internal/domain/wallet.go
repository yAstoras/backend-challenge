package domain

import (
	"errors"
	"time"
)

var (
	ErrInsufficientBalance = errors.New("insufficient balance")
	ErrInvalidWallet       = errors.New("invalid wallet")
	ErrInvalidMovement     = errors.New("invalid wallet movement")
)

type Wallet struct {
	ID        string
	PlayerID  string
	Currency  string
	Balance   Money
	Version   int64
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewWallet(
	id string,
	playerID string,
	initialBalance Money,
	now time.Time,
) (Wallet, error) {
	if id == "" {
		return Wallet{}, ErrInvalidWallet
	}

	if playerID == "" {
		return Wallet{}, ErrInvalidWallet
	}

	if initialBalance.Currency == "" {
		return Wallet{}, ErrInvalidWallet
	}

	if initialBalance.Amount < 0 {
		return Wallet{}, ErrInvalidWallet
	}

	if now.IsZero() {
		return Wallet{}, ErrInvalidWallet
	}

	return Wallet{
		ID:        id,
		PlayerID:  playerID,
		Currency:  initialBalance.Currency,
		Balance:   initialBalance,
		Version:   1,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

func (w *Wallet) Debit(amount Money) error {
	if amount.Amount <= 0 {
		return ErrInvalidMovement
	}

	if w.Balance.Currency != amount.Currency {
		return ErrCurrencyMismatch
	}

	if w.Balance.Amount < amount.Amount {
		return ErrInsufficientBalance
	}

	newBalance, err := w.Balance.Subtract(amount)

	if err != nil {
		return err
	}

	w.Balance = newBalance
	w.Version++
	w.UpdatedAt = time.Now()

	return nil
}

func (w *Wallet) Credit(amount Money) error {
	if amount.Amount <= 0 {
		return ErrInvalidMovement
	}

	if w.Balance.Currency != amount.Currency {
		return ErrCurrencyMismatch
	}

	newBalance, err := w.Balance.Add(amount)
	if err != nil {
		return err
	}

	w.Balance = newBalance
	w.Version++
	w.UpdatedAt = time.Now()

	return nil
}
