package application

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func WithPostgresTransaction(
	ctx context.Context,
	db *pgxpool.Pool,
	fn func(tx pgx.Tx) error,
) error {
	tx, err := db.Begin(ctx)
	if err != nil {
		return err
	}

	defer func() {
		_ = tx.Rollback(ctx)
	}()

	if err := fn(tx); err != nil {
		return err
	}

	return tx.Commit(ctx)
}
