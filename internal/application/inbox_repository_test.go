package application

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func TestInboxRepositoryCreateDeduplicatesMessage(t *testing.T) {
	ctx := context.Background()

	db, err := pgxpool.New(ctx, "postgres://jungle:jungle@localhost:5432/jungle")
	if err != nil {
		t.Fatalf("failed to connect to postgres: %v", err)
	}
	defer db.Close()

	repository := NewInboxRepository(db)

	messageID := "test-message-inbox-001"
	idempotencyKey := "test-idempotency-inbox-001"

	_, err = db.Exec(ctx, `
		DELETE FROM inbox_messages
		WHERE message_id = $1
	`, messageID)
	if err != nil {
		t.Fatalf("failed to clean test data: %v", err)
	}

	inserted, err := repository.Create(ctx, messageID, idempotencyKey)
	if err != nil {
		t.Fatalf("first create failed: %v", err)
	}

	if !inserted {
		t.Fatal("expected first message to be inserted")
	}

	inserted, err = repository.Create(ctx, messageID, idempotencyKey)
	if err != nil {
		t.Fatalf("second create failed: %v", err)
	}

	if inserted {
		t.Fatal("expected second message to be ignored as duplicate")
	}
}

func TestInboxRepositoryCreateTxRollback(t *testing.T) {
	ctx := context.Background()

	db, err := pgxpool.New(ctx, "postgres://jungle:jungle@localhost:5432/jungle")
	if err != nil {
		t.Fatalf("failed to connect to postgres: %v", err)
	}
	defer db.Close()

	repository := NewInboxRepository(db)

	messageID := "test-message-inbox-tx-001"
	idempotencyKey := "test-idempotency-inbox-tx-001"

	_, err = db.Exec(ctx, `
		DELETE FROM inbox_messages
		WHERE message_id = $1
	`, messageID)
	if err != nil {
		t.Fatalf("failed to clean test data: %v", err)
	}

	tx, err := db.Begin(ctx)
	if err != nil {
		t.Fatalf("failed to begin transaction: %v", err)
	}

	inserted, err := repository.CreateTx(
		ctx,
		tx,
		messageID,
		idempotencyKey,
	)
	if err != nil {
		t.Fatalf("failed to create inbox message: %v", err)
	}

	if !inserted {
		t.Fatal("expected inbox message to be inserted")
	}

	if err := tx.Rollback(ctx); err != nil {
		t.Fatalf("failed to rollback transaction: %v", err)
	}

	var count int

	err = db.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM inbox_messages
		WHERE message_id = $1
	`, messageID).Scan(&count)
	if err != nil {
		t.Fatalf("failed to check inbox message: %v", err)
	}

	if count != 0 {
		t.Fatalf("expected inbox message to be rolled back, found %d", count)
	}
}

func TestInboxRepositoryCreateTxCommit(t *testing.T) {
	ctx := context.Background()

	db, err := pgxpool.New(ctx, "postgres://jungle:jungle@localhost:5432/jungle")
	if err != nil {
		t.Fatalf("failed to connect to postgres: %v", err)
	}
	defer db.Close()

	repository := NewInboxRepository(db)

	messageID := "test-message-inbox-tx-002"
	idempotencyKey := "test-idempotency-inbox-tx-002"

	_, err = db.Exec(ctx, `
		DELETE FROM inbox_messages
		WHERE message_id = $1
	`, messageID)
	if err != nil {
		t.Fatalf("failed to clean test data: %v", err)
	}

	tx, err := db.Begin(ctx)
	if err != nil {
		t.Fatalf("failed to begin transaction: %v", err)
	}

	inserted, err := repository.CreateTx(
		ctx,
		tx,
		messageID,
		idempotencyKey,
	)
	if err != nil {
		t.Fatalf("failed to create inbox message: %v", err)
	}

	if !inserted {
		t.Fatal("expected inbox message to be inserted")
	}

	if err := tx.Commit(ctx); err != nil {
		t.Fatalf("failed to commit transaction: %v", err)
	}

	var count int

	err = db.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM inbox_messages
		WHERE message_id = $1
	`, messageID).Scan(&count)
	if err != nil {
		t.Fatalf("failed to check inbox message: %v", err)
	}

	if count != 1 {
		t.Fatalf("expected inbox message to be committed, found %d", count)
	}
}

func TestInboxRepositoryMarkProcessedTx(t *testing.T) {
	ctx := context.Background()

	db, err := pgxpool.New(ctx, "postgres://jungle:jungle@localhost:5432/jungle")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repository := NewInboxRepository(db)

	messageID := "test-mark-processed-message"

	_, err = db.Exec(ctx, `
		DELETE FROM inbox_messages
		WHERE message_id = $1
	`, messageID)
	if err != nil {
		t.Fatal(err)
	}

	tx, err := db.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}

	created, err := repository.CreateTx(
		ctx,
		tx,
		messageID,
		"test-mark-processed-idempotency",
	)
	if err != nil {
		_ = tx.Rollback(ctx)
		t.Fatal(err)
	}

	if !created {
		_ = tx.Rollback(ctx)
		t.Fatal("expected inbox message to be created")
	}

	if err := repository.MarkProcessedTx(ctx, tx, messageID); err != nil {
		_ = tx.Rollback(ctx)
		t.Fatal(err)
	}

	if err := tx.Commit(ctx); err != nil {
		t.Fatal(err)
	}

	var processedAt *time.Time

	err = db.QueryRow(ctx, `
		SELECT processed_at
		FROM inbox_messages
		WHERE message_id = $1
	`, messageID).Scan(&processedAt)
	if err != nil {
		t.Fatal(err)
	}

	if processedAt == nil {
		t.Fatal("expected processed_at to be set")
	}

	_, err = db.Exec(ctx, `
		DELETE FROM inbox_messages
		WHERE message_id = $1
	`, messageID)
	if err != nil {
		t.Fatal(err)
	}
}
