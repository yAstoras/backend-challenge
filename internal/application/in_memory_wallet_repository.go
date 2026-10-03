package application

import (
	"context"
	"errors"

	"github.com/yasto/backend-challenge/internal/domain"
)

var ErrWalletNotFound = errors.New("wallet not found")

type InMemoryWalletRepository struct {
	wallets map[string]domain.Wallet
}

func NewInMemoryWalletRepository() *InMemoryWalletRepository {
	return &InMemoryWalletRepository{
		wallets: make(map[string]domain.Wallet),
	}
}

func (r *InMemoryWalletRepository) FindByID(
	ctx context.Context,
	id string,
) (domain.Wallet, error) {
	wallet, exists := r.wallets[id]

	if !exists {
		return domain.Wallet{}, ErrWalletNotFound
	}

	return wallet, nil
}

func (r *InMemoryWalletRepository) Save(
	ctx context.Context,
	wallet domain.Wallet,
) error {
	r.wallets[wallet.ID] = wallet

	return nil
}

func (r *InMemoryWalletRepository) Create(ctx context.Context, wallet domain.Wallet) error {
	if _, exists := r.wallets[wallet.ID]; exists {
		return errors.New("wallet already exists")
	}

	r.wallets[wallet.ID] = wallet
	return nil
}
