package application

import (
	"context"

	"github.com/yasto/backend-challenge/internal/domain"
)

type InMemoryLedgerRepository struct {
	entries []domain.WalletLedgerEntry
}

func NewInMemoryLedgerRepository() *InMemoryLedgerRepository {
	return &InMemoryLedgerRepository{
		entries: make([]domain.WalletLedgerEntry, 0),
	}
}

func (r *InMemoryLedgerRepository) Save(
	ctx context.Context,
	entry domain.WalletLedgerEntry,
) error {
	r.entries = append(r.entries, entry)
	return nil
}

func (r *InMemoryLedgerRepository) Entries() []domain.WalletLedgerEntry {
	return append([]domain.WalletLedgerEntry(nil), r.entries...)
}
