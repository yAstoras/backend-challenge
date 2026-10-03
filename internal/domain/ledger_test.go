package domain

import (
	"testing"
	"time"
)

func TestWalletLedgerEntry(t *testing.T) {
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

	entry := WalletLedgerEntry{
		ID:            "ledger-123",
		WalletID:      "wallet-123",
		TransactionID: "transaction-123",
		Direction:     LedgerDebit,
		Amount:        NewMoney(3000, "BRL"),
		BalanceBefore: NewMoney(10000, "BRL"),
		BalanceAfter:  NewMoney(7000, "BRL"),
		CreatedAt:     now,
	}

	if entry.ID != "ledger-123" {
		t.Fatalf("unexpected ID: %s", entry.ID)
	}

	if entry.Direction != LedgerDebit {
		t.Fatalf("expected DEBIT, got %s", entry.Direction)
	}

	if entry.Amount.Amount != 3000 {
		t.Fatalf("expected amount 3000, got %d", entry.Amount.Amount)
	}

	if entry.BalanceBefore.Amount != 10000 {
		t.Fatalf("expected balance before 10000, got %d", entry.BalanceBefore.Amount)
	}

	if entry.BalanceAfter.Amount != 7000 {
		t.Fatalf("expected balance after 7000, got %d", entry.BalanceAfter.Amount)
	}
}

func TestNewWalletLedgerEntryRejectsEmptyID(t *testing.T) {
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

	_, err := NewWalletLedgerEntry(
		"",
		"wallet-123",
		"transaction-123",
		LedgerDebit,
		NewMoney(3000, "BRL"),
		NewMoney(10000, "BRL"),
		NewMoney(7000, "BRL"),
		now,
	)

	if err != ErrInvalidLedgerEntry {
		t.Fatalf("expected invalid ledger entry error, got %v", err)
	}
}

func TestNewWalletLedgerEntryRejectsInvalidDirection(t *testing.T) {
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

	_, err := NewWalletLedgerEntry(
		"ledger-123",
		"wallet-123",
		"transaction-123",
		"INVALID",
		NewMoney(3000, "BRL"),
		NewMoney(10000, "BRL"),
		NewMoney(7000, "BRL"),
		now,
	)

	if err != ErrInvalidLedgerEntry {
		t.Fatalf("expected invalid ledger entry error, got %v", err)
	}
}

func TestNewWalletLedgerEntryRejectsNonPositiveAmount(t *testing.T) {
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

	_, err := NewWalletLedgerEntry(
		"ledger-123",
		"wallet-123",
		"transaction-123",
		LedgerDebit,
		NewMoney(0, "BRL"),
		NewMoney(10000, "BRL"),
		NewMoney(10000, "BRL"),
		now,
	)

	if err != ErrInvalidLedgerEntry {
		t.Fatalf("expected invalid ledger entry error, got %v", err)
	}
}

func TestNewWalletLedgerEntryRejectsDifferentCurrencies(t *testing.T) {
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

	_, err := NewWalletLedgerEntry(
		"ledger-123",
		"wallet-123",
		"transaction-123",
		LedgerDebit,
		NewMoney(3000, "BRL"),
		NewMoney(10000, "BRL"),
		NewMoney(7000, "USD"),
		now,
	)

	if err != ErrCurrencyMismatch {
		t.Fatalf("expected currency mismatch error, got %v", err)
	}
}

func TestNewWalletLedgerEntryRejectsIncorrectDebitBalance(t *testing.T) {
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

	_, err := NewWalletLedgerEntry(
		"ledger-123",
		"wallet-123",
		"transaction-123",
		LedgerDebit,
		NewMoney(3000, "BRL"),
		NewMoney(10000, "BRL"),
		NewMoney(9000, "BRL"),
		now,
	)

	if err != ErrInvalidLedgerEntry {
		t.Fatalf("expected invalid ledger entry error, got %v", err)
	}
}

func TestNewWalletLedgerEntryRejectsNegativeBalance(t *testing.T) {
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

	_, err := NewWalletLedgerEntry(
		"ledger-123",
		"wallet-123",
		"transaction-123",
		LedgerDebit,
		NewMoney(3000, "BRL"),
		NewMoney(1000, "BRL"),
		NewMoney(-2000, "BRL"),
		now,
	)

	if err != ErrInvalidLedgerEntry {
		t.Fatalf("expected invalid ledger entry error, got %v", err)
	}
}
