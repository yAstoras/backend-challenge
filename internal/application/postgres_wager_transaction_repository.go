package application

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yasto/backend-challenge/internal/domain"
)

var (
	ErrWagerTransactionNotFound    = errors.New("wager transaction not found")
	ErrIdempotencyConflict         = errors.New("idempotency conflict")
	ErrProviderTransactionConflict = errors.New("provider transaction conflict")
)

type PostgresWagerTransactionRepository struct {
	db *pgxpool.Pool
}

func NewPostgresWagerTransactionRepository(
	db *pgxpool.Pool,
) *PostgresWagerTransactionRepository {
	return &PostgresWagerTransactionRepository{
		db: db,
	}
}

func (r *PostgresWagerTransactionRepository) Create(
	ctx context.Context,
	transaction domain.WagerTransaction,
) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO wager_transactions (
			id,
			provider_id,
			external_transaction_id,
			idempotency_key,
			payload_hash,
			player_id,
			wallet_id,
			round_id,
			game_id,
			transaction_type,
			amount,
			currency,
			reference_external_transaction_id,
			state,
			created_at,
			updated_at
		)
		VALUES (
			$1,
			$2,
			$3,
			$4,
			$5,
			$6,
			$7,
			$8,
			$9,
			$10,
			$11,
			$12,
			$13,
			$14,
			$15,
			$16
		)
	`,
		transaction.ID,
		transaction.ProviderID,
		transaction.ExternalTransactionID,
		transaction.IdempotencyKey,
		transaction.PayloadHash,
		transaction.PlayerID,
		transaction.WalletID,
		transaction.RoundID,
		transaction.GameID,
		transaction.Type,
		transaction.Amount.Amount,
		transaction.Amount.Currency,
		transaction.ReferenceExternalTransactionID,
		transaction.State,
		transaction.CreatedAt,
		transaction.UpdatedAt,
	)

	return err
}

func (r *PostgresWagerTransactionRepository) FindByID(
	ctx context.Context,
	id string,
) (domain.WagerTransaction, error) {
	var transaction domain.WagerTransaction

	err := r.db.QueryRow(ctx, `
		SELECT
			id,
			provider_id,
			external_transaction_id,
			idempotency_key,
			payload_hash,
			player_id,
			wallet_id,
			round_id,
			game_id,
			transaction_type,
			amount,
			currency,
			reference_external_transaction_id,
			state,
			COALESCE(failure_code, ''),
			created_at,
			updated_at
		FROM wager_transactions
		WHERE id = $1
	`, id).Scan(
		&transaction.ID,
		&transaction.ProviderID,
		&transaction.ExternalTransactionID,
		&transaction.IdempotencyKey,
		&transaction.PayloadHash,
		&transaction.PlayerID,
		&transaction.WalletID,
		&transaction.RoundID,
		&transaction.GameID,
		&transaction.Type,
		&transaction.Amount.Amount,
		&transaction.Amount.Currency,
		&transaction.ReferenceExternalTransactionID,
		&transaction.State,
		&transaction.FailureCode,
		&transaction.CreatedAt,
		&transaction.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return domain.WagerTransaction{}, ErrWagerTransactionNotFound
	}

	if err != nil {
		return domain.WagerTransaction{}, err
	}

	return transaction, nil
}

func (r *PostgresWagerTransactionRepository) FindByProviderExternalID(
	ctx context.Context,
	providerID string,
	externalTransactionID string,
) (domain.WagerTransaction, error) {
	var transaction domain.WagerTransaction

	err := r.db.QueryRow(ctx, `
		SELECT
			id,
			provider_id,
			external_transaction_id,
			idempotency_key,
			payload_hash,
			player_id,
			wallet_id,
			round_id,
			game_id,
			transaction_type,
			amount,
			currency,
			reference_external_transaction_id,
			state,
			created_at,
			updated_at
		FROM wager_transactions
		WHERE provider_id = $1
		AND external_transaction_id = $2
		`,
		providerID,
		externalTransactionID,
	).Scan(
		&transaction.ID,
		&transaction.ProviderID,
		&transaction.ExternalTransactionID,
		&transaction.IdempotencyKey,
		&transaction.PayloadHash,
		&transaction.PlayerID,
		&transaction.WalletID,
		&transaction.RoundID,
		&transaction.GameID,
		&transaction.Type,
		&transaction.Amount.Amount,
		&transaction.Amount.Currency,
		&transaction.ReferenceExternalTransactionID,
		&transaction.State,
		&transaction.CreatedAt,
		&transaction.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return domain.WagerTransaction{}, ErrWagerTransactionNotFound
	}

	if err != nil {
		return domain.WagerTransaction{}, err
	}

	return transaction, nil
}

func (r *PostgresWagerTransactionRepository) FindByIdempotencyKey(
	ctx context.Context,
	idempotencyKey string,
) (domain.WagerTransaction, error) {
	var transaction domain.WagerTransaction

	err := r.db.QueryRow(ctx, `
		SELECT
			id,
			provider_id,
			external_transaction_id,
			idempotency_key,
			payload_hash,
			player_id,
			wallet_id,
			round_id,
			game_id,
			transaction_type,
			amount,
			currency,
			reference_external_transaction_id,
			state,
			COALESCE(failure_code, ''),
			created_at,
			updated_at
		FROM wager_transactions
		WHERE idempotency_key = $1
	`, idempotencyKey).Scan(
		&transaction.ID,
		&transaction.ProviderID,
		&transaction.ExternalTransactionID,
		&transaction.IdempotencyKey,
		&transaction.PayloadHash,
		&transaction.PlayerID,
		&transaction.WalletID,
		&transaction.RoundID,
		&transaction.GameID,
		&transaction.Type,
		&transaction.Amount.Amount,
		&transaction.Amount.Currency,
		&transaction.ReferenceExternalTransactionID,
		&transaction.State,
		&transaction.FailureCode,
		&transaction.CreatedAt,
		&transaction.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return domain.WagerTransaction{}, ErrWagerTransactionNotFound
	}

	if err != nil {
		return domain.WagerTransaction{}, err
	}

	return transaction, nil
}

func (r *PostgresWagerTransactionRepository) Update(
	ctx context.Context,
	transaction domain.WagerTransaction,
) error {
	_, err := r.db.Exec(ctx, `
		UPDATE wager_transactions
		SET
			state = $1,
			updated_at = $2
		WHERE id = $3
	`,
		transaction.State,
		transaction.UpdatedAt,
		transaction.ID,
	)

	return err
}

func (r *PostgresWagerTransactionRepository) CreateTx(
	ctx context.Context,
	tx pgx.Tx,
	transaction domain.WagerTransaction,
) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO wager_transactions (
			id,
			provider_id,
			external_transaction_id,
			idempotency_key,
			payload_hash,
			player_id,
			wallet_id,
			round_id,
			game_id,
			transaction_type,
			amount,
			currency,
			reference_external_transaction_id,
			state,
			created_at,
			updated_at
		)
		VALUES (
			$1,
			$2,
			$3,
			$4,
			$5,
			$6,
			$7,
			$8,
			$9,
			$10,
			$11,
			$12,
			$13,
			$14,
			$15,
			$16
		)
	`,
		transaction.ID,
		transaction.ProviderID,
		transaction.ExternalTransactionID,
		transaction.IdempotencyKey,
		transaction.PayloadHash,
		transaction.PlayerID,
		transaction.WalletID,
		transaction.RoundID,
		transaction.GameID,
		transaction.Type,
		transaction.Amount.Amount,
		transaction.Amount.Currency,
		transaction.ReferenceExternalTransactionID,
		transaction.State,
		transaction.CreatedAt,
		transaction.UpdatedAt,
	)

	return err
}

func (r *PostgresWagerTransactionRepository) UpdateTx(
	ctx context.Context,
	tx pgx.Tx,
	transaction domain.WagerTransaction,
) error {
	_, err := tx.Exec(ctx, `
		UPDATE wager_transactions
		SET
			state = $1,
			failure_code = $2,
			updated_at = $3,
			reference_attempts = $4,
			next_reference_attempt_at = $5
		WHERE id = $6
	`,
		transaction.State,
		transaction.FailureCode,
		transaction.UpdatedAt,
		transaction.ReferenceAttempts,
		transaction.NextReferenceAttemptAt,
		transaction.ID,
	)

	return err
}

func (r *PostgresWagerTransactionRepository) FindByProviderExternalIDTx(
	ctx context.Context,
	tx pgx.Tx,
	providerID string,
	externalTransactionID string,
) (domain.WagerTransaction, error) {
	var transaction domain.WagerTransaction

	err := tx.QueryRow(ctx, `
		SELECT
			id,
			provider_id,
			external_transaction_id,
			idempotency_key,
			payload_hash,
			wallet_id,
			player_id,
			round_id,
			game_id,
			transaction_type,
			amount,
			currency,
			reference_external_transaction_id,
			state,
			created_at,
			updated_at
		FROM wager_transactions
		WHERE provider_id = $1
		  AND external_transaction_id = $2
		FOR UPDATE
	`,
		providerID,
		externalTransactionID,
	).Scan(
		&transaction.ID,
		&transaction.ProviderID,
		&transaction.ExternalTransactionID,
		&transaction.IdempotencyKey,
		&transaction.PayloadHash,
		&transaction.WalletID,
		&transaction.PlayerID,
		&transaction.RoundID,
		&transaction.GameID,
		&transaction.Type,
		&transaction.Amount.Amount,
		&transaction.Amount.Currency,
		&transaction.ReferenceExternalTransactionID,
		&transaction.State,
		&transaction.CreatedAt,
		&transaction.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return domain.WagerTransaction{}, ErrWagerTransactionNotFound
	}

	if err != nil {
		return domain.WagerTransaction{}, err
	}

	return transaction, nil
}

func (r *PostgresWagerTransactionRepository) FindProcessedReversalByReferenceTx(
	ctx context.Context,
	tx pgx.Tx,
	providerID string,
	referenceExternalTransactionID string,
) (domain.WagerTransaction, error) {
	var transaction domain.WagerTransaction

	err := tx.QueryRow(ctx, `
		SELECT
			id,
			provider_id,
			external_transaction_id,
			idempotency_key,
			payload_hash,
			wallet_id,
			player_id,
			round_id,
			game_id,
			transaction_type,
			amount,
			currency,
			reference_external_transaction_id,
			state,
			COALESCE(failure_code, ''),
			created_at,
			updated_at
		FROM wager_transactions
		WHERE provider_id = $1
		  AND reference_external_transaction_id = $2
		  AND transaction_type IN ('REFUND', 'ROLLBACK')
		  AND state = 'PROCESSED'
		LIMIT 1
	`,
		providerID,
		referenceExternalTransactionID,
	).Scan(
		&transaction.ID,
		&transaction.ProviderID,
		&transaction.ExternalTransactionID,
		&transaction.IdempotencyKey,
		&transaction.PayloadHash,
		&transaction.WalletID,
		&transaction.PlayerID,
		&transaction.RoundID,
		&transaction.GameID,
		&transaction.Type,
		&transaction.Amount.Amount,
		&transaction.Amount.Currency,
		&transaction.ReferenceExternalTransactionID,
		&transaction.State,
		&transaction.FailureCode,
		&transaction.CreatedAt,
		&transaction.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return domain.WagerTransaction{}, ErrWagerTransactionNotFound
	}

	if err != nil {
		return domain.WagerTransaction{}, err
	}

	return transaction, nil
}

func (r *PostgresWagerTransactionRepository) FindPendingReference(
	ctx context.Context,
	limit int,
) ([]domain.WagerTransaction, error) {
	if limit <= 0 {
		return nil, errors.New("limit must be positive")
	}

	rows, err := r.db.Query(ctx, `
		SELECT
			id,
			provider_id,
			external_transaction_id,
			idempotency_key,
			payload_hash,
			player_id,
			wallet_id,
			round_id,
			game_id,
			transaction_type,
			amount,
			currency,
			reference_external_transaction_id,
			state,
			created_at,
			updated_at,
			reference_attempts,
			next_reference_attempt_at
		FROM wager_transactions
		WHERE state = 'PENDING_REFERENCE'
		AND next_reference_attempt_at <= NOW()
		ORDER BY created_at ASC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []domain.WagerTransaction

	for rows.Next() {
		var transaction domain.WagerTransaction

		err := rows.Scan(
			&transaction.ID,
			&transaction.ProviderID,
			&transaction.ExternalTransactionID,
			&transaction.IdempotencyKey,
			&transaction.PayloadHash,
			&transaction.PlayerID,
			&transaction.WalletID,
			&transaction.RoundID,
			&transaction.GameID,
			&transaction.Type,
			&transaction.Amount.Amount,
			&transaction.Amount.Currency,
			&transaction.ReferenceExternalTransactionID,
			&transaction.State,
			&transaction.CreatedAt,
			&transaction.UpdatedAt,
			&transaction.ReferenceAttempts,
			&transaction.NextReferenceAttemptAt,
		)
		if err != nil {
			return nil, err
		}

		transactions = append(transactions, transaction)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return transactions, nil
}

func (r *PostgresWagerTransactionRepository) FindByIDTx(
	ctx context.Context,
	tx pgx.Tx,
	id string,
) (domain.WagerTransaction, error) {
	var transaction domain.WagerTransaction

	err := tx.QueryRow(ctx, `
		SELECT
			id,
			provider_id,
			external_transaction_id,
			idempotency_key,
			payload_hash,
			player_id,
			wallet_id,
			round_id,
			game_id,
			transaction_type,
			amount,
			currency,
			reference_external_transaction_id,
			state,
			COALESCE(failure_code, ''),
			created_at,
			updated_at
		FROM wager_transactions
		WHERE id = $1
		FOR UPDATE
	`, id).Scan(
		&transaction.ID,
		&transaction.ProviderID,
		&transaction.ExternalTransactionID,
		&transaction.IdempotencyKey,
		&transaction.PayloadHash,
		&transaction.PlayerID,
		&transaction.WalletID,
		&transaction.RoundID,
		&transaction.GameID,
		&transaction.Type,
		&transaction.Amount.Amount,
		&transaction.Amount.Currency,
		&transaction.ReferenceExternalTransactionID,
		&transaction.State,
		&transaction.FailureCode,
		&transaction.CreatedAt,
		&transaction.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.WagerTransaction{}, ErrWagerTransactionNotFound
		}
		return domain.WagerTransaction{}, err
	}

	return transaction, nil
}
