package application

import (
	"context"
	"testing"
	"time"

	"github.com/yasto/backend-challenge/internal/domain"
)

func TestInMemoryWalletRepository(t *testing.T) {
	repository := NewInMemoryWalletRepository()

	wallet, err := domain.NewWallet(
		"wallet-123",
		"player-123",
		domain.NewMoney(10000, "BRL"),
		time.Now(),
	)

	if err != nil {
		t.Fatalf("unexpected error creating wallet: %v", err)
	}

	err = repository.Save(context.Background(), wallet)

	if err != nil {
		t.Fatalf("unexpected error saving wallet: %v", err)
	}

	foundWallet, err := repository.FindByID(
		context.Background(),
		"wallet-123",
	)

	if err != nil {
		t.Fatalf("unexpected error finding wallet: %v", err)
	}

	if foundWallet.ID != wallet.ID {
		t.Fatalf(
			"expected wallet ID %s, got %s",
			wallet.ID,
			foundWallet.ID,
		)
	}

	if foundWallet.Balance.Amount != 10000 {
		t.Fatalf(
			"expected balance 10000, got %d",
			foundWallet.Balance.Amount,
		)
	}
}

func TestInMemoryWalletRepositoryWalletNotFound(t *testing.T) {
	repository := NewInMemoryWalletRepository()

	_, err := repository.FindByID(
		context.Background(),
		"wallet-does-not-exist",
	)

	if err != ErrWalletNotFound {
		t.Fatalf(
			"expected ErrWalletNotFound, got %v",
			err,
		)
	}
}
