package application

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yasto/backend-challenge/internal/domain"
)

var ErrWalletConcurrency = errors.New("wallet concurrency conflict")

type PostgresWalletRepository struct {
	db *pgxpool.Pool
}

func NewPostgresWalletRepository(db *pgxpool.Pool) *PostgresWalletRepository {
	return &PostgresWalletRepository{
		db: db,
	}
}

func (r *PostgresWalletRepository) FindByID(
	ctx context.Context,
	id string,
) (domain.Wallet, error) {
	var wallet domain.Wallet

	err := r.db.QueryRow(
		ctx,
		`
		SELECT
			id,
			player_id,
			currency,
			balance,
			version,
			created_at,
			updated_at
		FROM wallets
		WHERE id = $1
		`,
		id,
	).Scan(
		&wallet.ID,
		&wallet.PlayerID,
		&wallet.Currency,
		&wallet.Balance.Amount,
		&wallet.Version,
		&wallet.CreatedAt,
		&wallet.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Wallet{}, ErrWalletNotFound
	}

	if err != nil {
		return domain.Wallet{}, err
	}

	wallet.Balance.Currency = wallet.Currency

	return wallet, nil
}

func (r *PostgresWalletRepository) Save(ctx context.Context, wallet domain.Wallet) error {
	previousVersion := wallet.Version - 1

	result, err := r.db.Exec(
		ctx,
		`
		UPDATE wallets
		SET
			balance = $1,
			version = $2,
			updated_at = $3
		WHERE id = $4
		  AND version = $5
		`,
		wallet.Balance.Amount,
		wallet.Version,
		wallet.UpdatedAt,
		wallet.ID,
		previousVersion,
	)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrWalletConcurrency
	}

	return nil
}

func (r *PostgresWalletRepository) Create(ctx context.Context, wallet domain.Wallet) error {
	_, err := r.db.Exec(
		ctx,
		`
		INSERT INTO wallets (
			id,
			player_id,
			currency,
			balance,
			version,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		`,
		wallet.ID,
		wallet.PlayerID,
		wallet.Currency,
		wallet.Balance.Amount,
		wallet.Version,
		wallet.CreatedAt,
		wallet.UpdatedAt,
	)

	return err
}

func (r *PostgresWalletRepository) FindByIDForUpdate(
	ctx context.Context,
	tx pgx.Tx,
	id string,
) (domain.Wallet, error) {
	var wallet domain.Wallet

	err := tx.QueryRow(ctx, `
		SELECT id, player_id, currency, balance, version, created_at, updated_at
		FROM wallets
		WHERE id = $1
		FOR UPDATE
	`, id).Scan(
		&wallet.ID,
		&wallet.PlayerID,
		&wallet.Currency,
		&wallet.Balance.Amount,
		&wallet.Version,
		&wallet.CreatedAt,
		&wallet.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Wallet{}, ErrWalletNotFound
	}

	if err != nil {
		return domain.Wallet{}, err
	}

	wallet.Balance.Currency = wallet.Currency

	return wallet, nil
}

func (r *PostgresWalletRepository) SaveTx(
	ctx context.Context,
	tx pgx.Tx,
	wallet domain.Wallet,
) error {
	_, err := tx.Exec(ctx, `
		UPDATE wallets
		SET balance = $1,
			version = $2,
			updated_at = $3
		WHERE id = $4
	`,
		wallet.Balance.Amount,
		wallet.Version,
		wallet.UpdatedAt,
		wallet.ID,
	)

	return err
}
