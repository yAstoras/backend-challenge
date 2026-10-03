package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/yasto/backend-challenge/internal/domain"
)

func TestCreateWagerTransaction_IdempotencyReplay(t *testing.T) {
	ctx := context.Background()

	pool, err := NewPostgresPool()
	if err != nil {
		t.Fatalf("failed to connect to postgres: %v", err)
	}
	defer pool.Close()

	repository := NewPostgresWagerTransactionRepository(pool)
	useCase := NewCreateWagerTransaction(repository)

	now := time.Now().UTC()

	transaction := createTestWagerTransaction(
		"wager-idempotency-replay",
		"idempotency-replay",
		"hash-replay",
		now,
	)

	_, err = pool.Exec(ctx, `
		DELETE FROM wager_transactions
		WHERE id = $1
	`, transaction.ID)
	if err != nil {
		t.Fatalf("failed to clean test transaction: %v", err)
	}

	first, err := useCase.Execute(ctx, transaction)
	if err != nil {
		t.Fatalf("first execution failed: %v", err)
	}

	second, err := useCase.Execute(ctx, transaction)
	if err != nil {
		t.Fatalf("second execution failed: %v", err)
	}

	if second.ID != first.ID {
		t.Fatalf(
			"expected replay to return transaction %q, got %q",
			first.ID,
			second.ID,
		)
	}

	if second.PayloadHash != first.PayloadHash {
		t.Fatalf(
			"expected payload hash %q, got %q",
			first.PayloadHash,
			second.PayloadHash,
		)
	}
}

func TestCreateWagerTransaction_IdempotencyConflict(t *testing.T) {
	ctx := context.Background()

	pool, err := NewPostgresPool()
	if err != nil {
		t.Fatalf("failed to connect to postgres: %v", err)
	}
	defer pool.Close()

	repository := NewPostgresWagerTransactionRepository(pool)
	useCase := NewCreateWagerTransaction(repository)

	now := time.Now().UTC()

	firstTransaction := createTestWagerTransaction(
		"wager-idempotency-conflict-1",
		"idempotency-conflict",
		"hash-A",
		now,
	)

	secondTransaction := createTestWagerTransaction(
		"wager-idempotency-conflict-2",
		"idempotency-conflict",
		"hash-B",
		now,
	)

	_, err = pool.Exec(ctx, `
		DELETE FROM wager_transactions
		WHERE idempotency_key = $1
	`, firstTransaction.IdempotencyKey)
	if err != nil {
		t.Fatalf("failed to clean test transactions: %v", err)
	}

	_, err = useCase.Execute(ctx, firstTransaction)
	if err != nil {
		t.Fatalf("first execution failed: %v", err)
	}

	_, err = useCase.Execute(ctx, secondTransaction)

	if !errors.Is(err, ErrIdempotencyConflict) {
		t.Fatalf(
			"expected ErrIdempotencyConflict, got %v",
			err,
		)
	}
}

func createTestWagerTransaction(
	id string,
	idempotencyKey string,
	payloadHash string,
	now time.Time,
) domain.WagerTransaction {
	transaction, err := domain.NewWagerTransaction(
		id,
		"provider-idempotency-test",
		id+"-external",
		idempotencyKey,
		payloadHash,
		"player-idempotency-test",
		"wallet-idempotency-test",
		"round-idempotency-test",
		"game-idempotency-test",
		domain.TransactionBet,
		domain.NewMoney(1000, "BRL"),
		"",
		now,
	)
	if err != nil {
		panic(err)
	}

	return transaction
}

func TestCreateWagerTransaction_ProviderExternalTransactionConflict(t *testing.T) {
	ctx := context.Background()

	pool, err := NewPostgresPool()
	if err != nil {
		t.Fatalf("failed to connect to postgres: %v", err)
	}
	defer pool.Close()

	repository := NewPostgresWagerTransactionRepository(pool)
	useCase := NewCreateWagerTransaction(repository)

	now := time.Now().UTC()

	firstTransaction := createTestWagerTransaction(
		"wager-provider-conflict-1",
		"idempotency-provider-1",
		"hash-provider-1",
		now,
	)

	secondTransaction := createTestWagerTransaction(
		"wager-provider-conflict-2",
		"idempotency-provider-2",
		"hash-provider-2",
		now,
	)

	// As duas transações precisam ter o mesmo provider + external ID.
	secondTransaction.ProviderID = firstTransaction.ProviderID
	secondTransaction.ExternalTransactionID = firstTransaction.ExternalTransactionID

	_, err = pool.Exec(ctx, `
		DELETE FROM wager_transactions
		WHERE provider_id = $1
		  AND external_transaction_id = $2
	`,
		firstTransaction.ProviderID,
		firstTransaction.ExternalTransactionID,
	)
	if err != nil {
		t.Fatalf("failed to clean test transaction: %v", err)
	}

	_, err = useCase.Execute(ctx, firstTransaction)
	if err != nil {
		t.Fatalf("first execution failed: %v", err)
	}

	_, err = useCase.Execute(ctx, secondTransaction)

	if !errors.Is(err, ErrProviderTransactionConflict) {
		t.Fatalf(
			"expected ErrProviderTransactionConflict, got %v",
			err,
		)
	}
}
