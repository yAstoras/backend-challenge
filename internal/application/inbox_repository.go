package application

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type InboxRepository struct {
	db *pgxpool.Pool
}

func NewInboxRepository(db *pgxpool.Pool) *InboxRepository {
	return &InboxRepository{
		db: db,
	}
}

func (r *InboxRepository) Create(
	ctx context.Context,
	messageID string,
	idempotencyKey string,
) (bool, error) {
	result, err := r.db.Exec(ctx, `
		INSERT INTO inbox_messages (
			message_id,
			idempotency_key
		)
		VALUES ($1, $2)
		ON CONFLICT DO NOTHING
	`, messageID, idempotencyKey)

	if err != nil {
		return false, fmt.Errorf("create inbox message: %w", err)
	}

	return result.RowsAffected() == 1, nil
}

func (r *InboxRepository) CreateTx(
	ctx context.Context,
	tx pgx.Tx,
	messageID string,
	idempotencyKey string,
) (bool, error) {
	result, err := tx.Exec(ctx, `
		INSERT INTO inbox_messages (
			message_id,
			idempotency_key
		)
		VALUES ($1, $2)
		ON CONFLICT DO NOTHING
	`, messageID, idempotencyKey)

	if err != nil {
		return false, fmt.Errorf("create inbox message in transaction: %w", err)
	}

	return result.RowsAffected() == 1, nil
}

func (r *InboxRepository) MarkProcessedTx(
	ctx context.Context,
	tx pgx.Tx,
	messageID string,
) error {
	_, err := tx.Exec(ctx, `
		UPDATE inbox_messages
		SET processed_at = NOW()
		WHERE message_id = $1
	`, messageID)
	if err != nil {
		return fmt.Errorf("mark inbox message as processed: %w", err)
	}

	return nil
}
