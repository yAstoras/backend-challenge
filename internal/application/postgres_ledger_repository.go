package application

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yasto/backend-challenge/internal/domain"
)

type PostgresLedgerRepository struct {
	db *pgxpool.Pool
}

func NewPostgresLedgerRepository(db *pgxpool.Pool) *PostgresLedgerRepository {
	return &PostgresLedgerRepository{
		db: db,
	}
}

func (r *PostgresLedgerRepository) Save(
	ctx context.Context,
	entry domain.WalletLedgerEntry,
) error {
	_, err := r.db.Exec(
		ctx,
		`
		INSERT INTO wallet_ledger (
			id,
			wallet_id,
			transaction_id,
			direction,
			amount,
			balance_before,
			balance_after,
			created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		`,
		entry.ID,
		entry.WalletID,
		entry.TransactionID,
		entry.Direction,
		entry.Amount.Amount,
		entry.BalanceBefore.Amount,
		entry.BalanceAfter.Amount,
		entry.CreatedAt,
	)

	return err
}

func (r *PostgresLedgerRepository) SaveTx(
	ctx context.Context,
	tx pgx.Tx,
	entry domain.WalletLedgerEntry,
) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO wallet_ledger (
			id,
			wallet_id,
			transaction_id,
			direction,
			amount,
			balance_before,
			balance_after,
			created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`,
		entry.ID,
		entry.WalletID,
		entry.TransactionID,
		entry.Direction,
		entry.Amount.Amount,
		entry.BalanceBefore.Amount,
		entry.BalanceAfter.Amount,
		entry.CreatedAt,
	)

	return err
}
