package application

import (
	"context"
	"testing"
	"time"

	"github.com/yasto/backend-challenge/internal/domain"
)

func TestPostgresWalletRepository_CreateAndFindByID(t *testing.T) {
	ctx := context.Background()

	pool, err := NewPostgresPool()
	if err != nil {
		t.Fatalf("failed to connect to postgres: %v", err)
	}
	defer pool.Close()

	repository := NewPostgresWalletRepository(pool)

	now := time.Now().UTC()

	wallet, err := domain.NewWallet(
		"wallet-test-1",
		"player-test-1",
		domain.NewMoney(10000, "BRL"),
		now,
	)
	if err != nil {
		t.Fatalf("failed to create wallet: %v", err)
	}

	_, err = pool.Exec(
		ctx,
		"DELETE FROM wallets WHERE id = $1",
		wallet.ID,
	)
	if err != nil {
		t.Fatalf("failed to clean test wallet: %v", err)
	}

	err = repository.Create(ctx, wallet)
	if err != nil {
		t.Fatalf("failed to create wallet in postgres: %v", err)
	}

	foundWallet, err := repository.FindByID(ctx, wallet.ID)
	if err != nil {
		t.Fatalf("failed to find wallet: %v", err)
	}

	if foundWallet.ID != wallet.ID {
		t.Fatalf(
			"expected wallet ID %s, got %s",
			wallet.ID,
			foundWallet.ID,
		)
	}

	if foundWallet.PlayerID != wallet.PlayerID {
		t.Fatalf(
			"expected player ID %s, got %s",
			wallet.PlayerID,
			foundWallet.PlayerID,
		)
	}

	if foundWallet.Currency != "BRL" {
		t.Fatalf(
			"expected currency BRL, got %s",
			foundWallet.Currency,
		)
	}

	if foundWallet.Balance.Amount != 10000 {
		t.Fatalf(
			"expected balance 10000, got %d",
			foundWallet.Balance.Amount,
		)
	}

	if foundWallet.Version != 1 {
		t.Fatalf(
			"expected version 1, got %d",
			foundWallet.Version,
		)
	}
}
