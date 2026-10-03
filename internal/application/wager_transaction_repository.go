package application

import (
	"context"

	"github.com/yasto/backend-challenge/internal/domain"
)

type WagerTransactionRepository interface {
	Create(ctx context.Context, transaction domain.WagerTransaction) error
	Update(ctx context.Context, transaction domain.WagerTransaction) error
	FindByID(ctx context.Context, id string) (domain.WagerTransaction, error)
	FindByProviderExternalID(ctx context.Context, providerID string, externalTransactionID string) (domain.WagerTransaction, error)
	FindByIdempotencyKey(ctx context.Context, idempotencyKey string) (domain.WagerTransaction, error)
}
