package domain

import (
	"testing"
	"time"
)

func TestWalletInsufficientBalance(t *testing.T) {
	now := time.Date(
		2026,
		time.January,
		1,
		12,
		0,
		0,
		0,
		time.UTC,
	)

	wallet, err := NewWallet(
		"wallet-123",
		"player-123",
		NewMoney(10000, "BRL"),
		now,
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = wallet.Debit(NewMoney(15000, "BRL"))

	if err != ErrInsufficientBalance {
		t.Fatalf("expected insufficient balance error, got %v", err)
	}

	if wallet.Balance.Amount != 10000 {
		t.Fatalf("expected balance to remain 10000, got %d", wallet.Balance.Amount)
	}
}

func TestWalletDebit(t *testing.T) {
	now := time.Date(
		2026,
		time.January,
		1,
		12,
		0,
		0,
		0,
		time.UTC,
	)

	wallet, err := NewWallet(
		"wallet-123",
		"player-123",
		NewMoney(10000, "BRL"),
		now,
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = wallet.Debit(NewMoney(3000, "BRL"))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if wallet.Balance.Amount != 7000 {
		t.Fatalf("expected 7000, got %d", wallet.Balance.Amount)
	}

	if wallet.Version != 2 {
		t.Fatalf("expected version 2, got %d", wallet.Version)
	}
}

func TestWalletCredit(t *testing.T) {
	now := time.Date(
		2026,
		time.January,
		1,
		12,
		0,
		0,
		0,
		time.UTC,
	)

	wallet, err := NewWallet(
		"wallet-123",
		"player-123",
		NewMoney(10000, "BRL"),
		now,
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = wallet.Credit(NewMoney(5000, "BRL"))

	if wallet.Balance.Amount != 15000 {
		t.Fatalf("expected 15000, got %d", wallet.Balance.Amount)
	}

	if wallet.Version != 2 {
		t.Fatalf("expected version 2, got %d", wallet.Version)
	}
}

func TestWalletDifferentCurrencies(t *testing.T) {
	now := time.Date(
		2026,
		time.January,
		1,
		12,
		0,
		0,
		0,
		time.UTC,
	)

	wallet, err := NewWallet(
		"wallet-123",
		"player-123",
		NewMoney(10000, "BRL"),
		now,
	)

	err = wallet.Credit(NewMoney(5000, "USD"))

	if err != ErrCurrencyMismatch {
		t.Fatalf("expected currency mismatch error, got %v", err)
	}

	if wallet.Balance.Amount != 10000 {
		t.Fatalf("expected balance to remain 10000, got %d", wallet.Balance.Amount)
	}
}

func TestNewWalletRejectsEmptyPlayerID(t *testing.T) {
	now := time.Date(
		2026,
		time.January,
		1,
		12,
		0,
		0,
		0,
		time.UTC,
	)

	_, err := NewWallet(
		"wallet-123",
		"",
		NewMoney(10000, "BRL"),
		now,
	)

	if err != ErrInvalidWallet {
		t.Fatalf("expected invalid wallet error, got %v", err)
	}
}

func TestNewWalletRejectsNegativeBalance(t *testing.T) {
	now := time.Date(
		2026,
		time.January,
		1,
		12,
		0,
		0,
		0,
		time.UTC,
	)

	_, err := NewWallet(
		"wallet-123",
		"player-123",
		NewMoney(-1000, "BRL"),
		now,
	)

	if err != ErrInvalidWallet {
		t.Fatalf("expected invalid wallet error, got %v", err)
	}
}

func TestWalletRejectsNegativeDebit(t *testing.T) {
	now := time.Date(
		2026,
		time.January,
		1,
		12,
		0,
		0,
		0,
		time.UTC,
	)

	wallet, err := NewWallet(
		"wallet-123",
		"player-123",
		NewMoney(10000, "BRL"),
		now,
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = wallet.Debit(NewMoney(-500, "BRL"))

	if err != ErrInvalidMovement {
		t.Fatalf("expected invalid movement error, got %v", err)
	}

	if wallet.Balance.Amount != 10000 {
		t.Fatalf("expected balance to remain 10000, got %d", wallet.Balance.Amount)
	}
}

func TestWalletRejectsNegativeCredit(t *testing.T) {
	now := time.Date(
		2026,
		time.January,
		1,
		12,
		0,
		0,
		0,
		time.UTC,
	)

	wallet, err := NewWallet(
		"wallet-123",
		"player-123",
		NewMoney(10000, "BRL"),
		now,
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = wallet.Credit(NewMoney(-500, "BRL"))

	if err != ErrInvalidMovement {
		t.Fatalf("expected invalid movement error, got %v", err)
	}

	if wallet.Balance.Amount != 10000 {
		t.Fatalf("expected balance to remain 10000, got %d", wallet.Balance.Amount)
	}
}
