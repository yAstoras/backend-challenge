package application

import (
	"context"
	"testing"
	"time"

	"github.com/yasto/backend-challenge/internal/domain"
)

func TestPostgresWagerTransactionRepository_CreateAndFindByID(t *testing.T) {
	ctx := context.Background()

	pool, err := NewPostgresPool()
	if err != nil {
		t.Fatalf("failed to connect to postgres: %v", err)
	}
	defer pool.Close()

	repository := NewPostgresWagerTransactionRepository(pool)

	now := time.Now().UTC()

	transaction, err := domain.NewWagerTransaction(
		"wager-repository-test-1",
		"provider-test",
		"external-test-1",
		"idempotency-test-1",
		"hash-test-1",
		"player-test-1",
		"wallet-test-1",
		"round-test-1",
		"game-test-1",
		domain.TransactionBet,
		domain.NewMoney(5000, "BRL"),
		"",
		now,
	)
	if err != nil {
		t.Fatalf("failed to create transaction: %v", err)
	}

	_, err = pool.Exec(ctx, `
		DELETE FROM wager_transactions
		WHERE id = $1
	`, transaction.ID)
	if err != nil {
		t.Fatalf("failed to clean test transaction: %v", err)
	}

	err = repository.Create(ctx, transaction)
	if err != nil {
		t.Fatalf("failed to create transaction in postgres: %v", err)
	}

	found, err := repository.FindByID(ctx, transaction.ID)
	if err != nil {
		t.Fatalf("failed to find transaction: %v", err)
	}

	if found.ID != transaction.ID {
		t.Fatalf("expected ID %s, got %s", transaction.ID, found.ID)
	}

	if found.ProviderID != transaction.ProviderID {
		t.Fatalf(
			"expected provider ID %s, got %s",
			transaction.ProviderID,
			found.ProviderID,
		)
	}

	if found.ExternalTransactionID != transaction.ExternalTransactionID {
		t.Fatalf(
			"expected external transaction ID %s, got %s",
			transaction.ExternalTransactionID,
			found.ExternalTransactionID,
		)
	}

	if found.IdempotencyKey != transaction.IdempotencyKey {
		t.Fatalf(
			"expected idempotency key %s, got %s",
			transaction.IdempotencyKey,
			found.IdempotencyKey,
		)
	}

	if found.PayloadHash != transaction.PayloadHash {
		t.Fatalf(
			"expected payload hash %s, got %s",
			transaction.PayloadHash,
			found.PayloadHash,
		)
	}

	if found.PlayerID != transaction.PlayerID {
		t.Fatalf(
			"expected player ID %s, got %s",
			transaction.PlayerID,
			found.PlayerID,
		)
	}

	if found.WalletID != transaction.WalletID {
		t.Fatalf(
			"expected wallet ID %s, got %s",
			transaction.WalletID,
			found.WalletID,
		)
	}

	if found.RoundID != transaction.RoundID {
		t.Fatalf(
			"expected round ID %s, got %s",
			transaction.RoundID,
			found.RoundID,
		)
	}

	if found.GameID != transaction.GameID {
		t.Fatalf(
			"expected game ID %s, got %s",
			transaction.GameID,
			found.GameID,
		)
	}

	if found.Type != transaction.Type {
		t.Fatalf(
			"expected type %s, got %s",
			transaction.Type,
			found.Type,
		)
	}

	if found.Amount.Amount != transaction.Amount.Amount {
		t.Fatalf(
			"expected amount %d, got %d",
			transaction.Amount.Amount,
			found.Amount.Amount,
		)
	}

	if found.Amount.Currency != transaction.Amount.Currency {
		t.Fatalf(
			"expected currency %s, got %s",
			transaction.Amount.Currency,
			found.Amount.Currency,
		)
	}

	if found.State != transaction.State {
		t.Fatalf(
			"expected state %s, got %s",
			transaction.State,
			found.State,
		)
	}
}

func TestPostgresWagerTransactionRepository_FindByProviderExternalID(t *testing.T) {
	ctx := context.Background()

	pool, err := NewPostgresPool()
	if err != nil {
		t.Fatalf("failed to connect to postgres: %v", err)
	}
	defer pool.Close()

	repository := NewPostgresWagerTransactionRepository(pool)

	now := time.Now().UTC()

	transaction, err := domain.NewWagerTransaction(
		"wager-repository-test-2",
		"provider-test-2",
		"external-test-2",
		"idempotency-test-2",
		"hash-test-2",
		"player-test-2",
		"wallet-test-2",
		"round-test-2",
		"game-test-2",
		domain.TransactionBet,
		domain.NewMoney(7500, "BRL"),
		"",
		now,
	)
	if err != nil {
		t.Fatalf("failed to create domain transaction: %v", err)
	}

	_, err = pool.Exec(ctx, `
		DELETE FROM wager_transactions
		WHERE id = $1
	`, transaction.ID)
	if err != nil {
		t.Fatalf("failed to clean test transaction: %v", err)
	}

	err = repository.Create(ctx, transaction)
	if err != nil {
		t.Fatalf("failed to create transaction in postgres: %v", err)
	}

	found, err := repository.FindByProviderExternalID(
		ctx,
		"provider-test-2",
		"external-test-2",
	)
	if err != nil {
		t.Fatalf("failed to find transaction: %v", err)
	}

	if found.ID != transaction.ID {
		t.Fatalf("expected ID %q, got %q", transaction.ID, found.ID)
	}

	if found.ProviderID != transaction.ProviderID {
		t.Fatalf("expected provider ID %q, got %q", transaction.ProviderID, found.ProviderID)
	}

	if found.ExternalTransactionID != transaction.ExternalTransactionID {
		t.Fatalf(
			"expected external transaction ID %q, got %q",
			transaction.ExternalTransactionID,
			found.ExternalTransactionID,
		)
	}

	if found.Amount.Amount != transaction.Amount.Amount {
		t.Fatalf(
			"expected amount %d, got %d",
			transaction.Amount.Amount,
			found.Amount.Amount,
		)
	}
}

func TestPostgresWagerTransactionRepository_FindByIdempotencyKey(t *testing.T) {
	ctx := context.Background()

	pool, err := NewPostgresPool()
	if err != nil {
		t.Fatalf("failed to connect to postgres: %v", err)
	}
	defer pool.Close()

	repository := NewPostgresWagerTransactionRepository(pool)

	now := time.Now().UTC()

	transaction, err := domain.NewWagerTransaction(
		"wager-repository-test-3",
		"provider-test-3",
		"external-test-3",
		"idempotency-test-3",
		"hash-test-3",
		"player-test-3",
		"wallet-test-3",
		"round-test-3",
		"game-test-3",
		domain.TransactionBet,
		domain.NewMoney(9000, "BRL"),
		"",
		now,
	)
	if err != nil {
		t.Fatalf("failed to create domain transaction: %v", err)
	}

	_, err = pool.Exec(ctx, `
		DELETE FROM wager_transactions
		WHERE id = $1
	`, transaction.ID)
	if err != nil {
		t.Fatalf("failed to clean test transaction: %v", err)
	}

	err = repository.Create(ctx, transaction)
	if err != nil {
		t.Fatalf("failed to create transaction in postgres: %v", err)
	}

	found, err := repository.FindByIdempotencyKey(
		ctx,
		transaction.IdempotencyKey,
	)
	if err != nil {
		t.Fatalf("failed to find transaction: %v", err)
	}

	if found.ID != transaction.ID {
		t.Fatalf("expected ID %q, got %q", transaction.ID, found.ID)
	}

	if found.IdempotencyKey != transaction.IdempotencyKey {
		t.Fatalf(
			"expected idempotency key %q, got %q",
			transaction.IdempotencyKey,
			found.IdempotencyKey,
		)
	}

	if found.PayloadHash != transaction.PayloadHash {
		t.Fatalf(
			"expected payload hash %q, got %q",
			transaction.PayloadHash,
			found.PayloadHash,
		)
	}

	if found.Amount.Amount != transaction.Amount.Amount {
		t.Fatalf(
			"expected amount %d, got %d",
			transaction.Amount.Amount,
			found.Amount.Amount,
		)
	}
}

func TestPostgresWagerTransactionRepository_Update(t *testing.T) {
	ctx := context.Background()

	pool, err := NewPostgresPool()
	if err != nil {
		t.Fatalf("failed to connect to postgres: %v", err)
	}
	defer pool.Close()

	repository := NewPostgresWagerTransactionRepository(pool)

	now := time.Now().UTC()

	transaction, err := domain.NewWagerTransaction(
		"wager-repository-update-test",
		"provider-update-test",
		"external-update-test",
		"idempotency-update-test",
		"hash-update-test",
		"player-update-test",
		"wallet-update-test",
		"round-update-test",
		"game-update-test",
		domain.TransactionBet,
		domain.NewMoney(5000, "BRL"),
		"",
		now,
	)
	if err != nil {
		t.Fatalf("failed to create domain transaction: %v", err)
	}

	_, err = pool.Exec(ctx, `
		DELETE FROM wager_transactions
		WHERE id = $1
	`, transaction.ID)
	if err != nil {
		t.Fatalf("failed to clean test transaction: %v", err)
	}

	err = repository.Create(ctx, transaction)
	if err != nil {
		t.Fatalf("failed to create transaction: %v", err)
	}

	if err := transaction.StartProcessing(); err != nil {
		t.Fatalf("failed to start processing: %v", err)
	}

	if err := transaction.Process(); err != nil {
		t.Fatalf("failed to process transaction: %v", err)
	}

	err = repository.Update(ctx, transaction)
	if err != nil {
		t.Fatalf("failed to update transaction: %v", err)
	}

	found, err := repository.FindByID(ctx, transaction.ID)
	if err != nil {
		t.Fatalf("failed to find updated transaction: %v", err)
	}

	if found.State != transaction.State {
		t.Fatalf(
			"expected state %q, got %q",
			transaction.State,
			found.State,
		)
	}
}
