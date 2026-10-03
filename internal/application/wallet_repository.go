package application

import (
	"context"

	"github.com/yasto/backend-challenge/internal/domain"
)

type WalletRepository interface {
	FindByID(ctx context.Context, id string) (domain.Wallet, error)
	Save(ctx context.Context, wallet domain.Wallet) error
	Create(ctx context.Context, wallet domain.Wallet) error
}
