package application

import (
	"context"

	"github.com/yasto/backend-challenge/internal/domain"
)

type LedgerRepository interface {
	Save(ctx context.Context, entry domain.WalletLedgerEntry) error
}
