package application

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yasto/backend-challenge/internal/domain"
)

func setupPostgresWagerTransactionProcessorTest(t *testing.T) (
	context.Context,
	*pgxpool.Pool,
	*PostgresWagerTransactionProcessor,
) {
	t.Helper()

	ctx := context.Background()

	pool, err := NewPostgresPool()
	if err != nil {
		t.Fatalf("failed to create postgres pool: %v", err)
	}

	t.Cleanup(func() {
		pool.Close()
	})

	walletRepository := NewPostgresWalletRepository(pool)
	ledgerRepository := NewPostgresLedgerRepository(pool)
	wagerTransactionRepository := NewPostgresWagerTransactionRepository(pool)
	outboxRepository := NewOutboxRepository(pool, nil)

	processor := NewPostgresWagerTransactionProcessor(
		pool,
		walletRepository,
		ledgerRepository,
		wagerTransactionRepository,
		outboxRepository,
	)

	return ctx, pool, processor
}

func TestPostgresWagerTransactionProcessor_RefundBet(t *testing.T) {
	ctx, pool, processor := setupPostgresWagerTransactionProcessorTest(t)

	now := time.Now()

	_, err := pool.Exec(ctx, `
		DELETE FROM wallet_ledger
		WHERE wallet_id = 'wallet-refund-test'
	`)
	if err != nil {
		t.Fatalf("failed to clean ledger: %v", err)
	}

	_, err = pool.Exec(ctx, `
		DELETE FROM wager_transactions
		WHERE id IN ('bet-refund-test', 'refund-test')
	`)
	if err != nil {
		t.Fatalf("failed to clean wager transactions: %v", err)
	}

	_, err = pool.Exec(ctx, `
		DELETE FROM wallets
		WHERE id = 'wallet-refund-test'
	`)
	if err != nil {
		t.Fatalf("failed to clean wallet: %v", err)
	}

	wallet, err := domain.NewWallet(
		"wallet-refund-test",
		"player-refund-test",
		domain.NewMoney(10000, "BRL"),
		now,
	)
	if err != nil {
		t.Fatalf("failed to create wallet: %v", err)
	}

	walletRepository := NewPostgresWalletRepository(pool)

	if err := walletRepository.Create(ctx, wallet); err != nil {
		t.Fatalf("failed to persist wallet: %v", err)
	}

	bet, err := domain.NewWagerTransaction(
		"bet-refund-test",
		"provider-test",
		"external-bet-refund-test",
		"idempotency-bet-refund-test",
		"hash-bet-refund-test",
		"player-refund-test",
		"wallet-refund-test",
		"round-refund-test",
		"game-refund-test",
		domain.TransactionBet,
		domain.NewMoney(3000, "BRL"),
		"",
		now,
	)
	if err != nil {
		t.Fatalf("failed to create bet: %v", err)
	}

	if err := processor.Execute(ctx, bet); err != nil {
		t.Fatalf("failed to process bet: %v", err)
	}

	refund, err := domain.NewWagerTransaction(
		"refund-test",
		"provider-test",
		"external-refund-test",
		"idempotency-refund-test",
		"hash-refund-test",
		"player-refund-test",
		"wallet-refund-test",
		"round-refund-test",
		"game-refund-test",
		domain.TransactionRefund,
		domain.NewMoney(3000, "BRL"),
		"external-bet-refund-test",
		now.Add(time.Second),
	)
	if err != nil {
		t.Fatalf("failed to create refund: %v", err)
	}

	if err := processor.Execute(ctx, refund); err != nil {
		t.Fatalf("failed to process refund: %v", err)
	}

	result, err := walletRepository.FindByID(ctx, "wallet-refund-test")
	if err != nil {
		t.Fatalf("failed to find wallet: %v", err)
	}

	if result.Balance.Amount != 10000 {
		t.Fatalf("expected final balance 10000, got %d", result.Balance.Amount)
	}

	if result.Version != 3 {
		t.Fatalf("expected wallet version 3, got %d", result.Version)
	}
}

func TestPostgresWagerTransactionProcessor_RollbackBet(t *testing.T) {
	ctx, pool, processor := setupPostgresWagerTransactionProcessorTest(t)

	now := time.Now()

	_, err := pool.Exec(ctx, `
		DELETE FROM wallet_ledger
		WHERE wallet_id = 'wallet-rollback-bet-test'
	`)
	if err != nil {
		t.Fatalf("failed to clean ledger: %v", err)
	}

	_, err = pool.Exec(ctx, `
		DELETE FROM wager_transactions
		WHERE id IN ('bet-rollback-test', 'rollback-bet-test')
	`)
	if err != nil {
		t.Fatalf("failed to clean wager transactions: %v", err)
	}

	_, err = pool.Exec(ctx, `
		DELETE FROM wallets
		WHERE id = 'wallet-rollback-bet-test'
	`)
	if err != nil {
		t.Fatalf("failed to clean wallet: %v", err)
	}

	wallet, err := domain.NewWallet(
		"wallet-rollback-bet-test",
		"player-rollback-bet-test",
		domain.NewMoney(10000, "BRL"),
		now,
	)
	if err != nil {
		t.Fatalf("failed to create wallet: %v", err)
	}

	walletRepository := NewPostgresWalletRepository(pool)

	if err := walletRepository.Create(ctx, wallet); err != nil {
		t.Fatalf("failed to persist wallet: %v", err)
	}

	bet, err := domain.NewWagerTransaction(
		"bet-rollback-test",
		"provider-rollback-test",
		"external-bet-rollback-test",
		"idempotency-bet-rollback-test",
		"hash-bet-rollback-test",
		"player-rollback-bet-test",
		"wallet-rollback-bet-test",
		"round-rollback-test",
		"game-rollback-test",
		domain.TransactionBet,
		domain.NewMoney(3000, "BRL"),
		"",
		now,
	)
	if err != nil {
		t.Fatalf("failed to create bet: %v", err)
	}

	if err := processor.Execute(ctx, bet); err != nil {
		t.Fatalf("failed to process bet: %v", err)
	}

	rollback, err := domain.NewWagerTransaction(
		"rollback-bet-test",
		"provider-rollback-test",
		"external-rollback-bet-test",
		"idempotency-rollback-bet-test",
		"hash-rollback-bet-test",
		"player-rollback-bet-test",
		"wallet-rollback-bet-test",
		"round-rollback-test",
		"game-rollback-test",
		domain.TransactionRollback,
		domain.NewMoney(3000, "BRL"),
		"external-bet-rollback-test",
		now.Add(time.Second),
	)
	if err != nil {
		t.Fatalf("failed to create rollback: %v", err)
	}

	if err := processor.Execute(ctx, rollback); err != nil {
		t.Fatalf("failed to process rollback: %v", err)
	}

	result, err := walletRepository.FindByID(ctx, "wallet-rollback-bet-test")
	if err != nil {
		t.Fatalf("failed to find wallet: %v", err)
	}

	if result.Balance.Amount != 10000 {
		t.Fatalf("expected final balance 10000, got %d", result.Balance.Amount)
	}

	if result.Version != 3 {
		t.Fatalf("expected wallet version 3, got %d", result.Version)
	}

	var direction string

	err = pool.QueryRow(ctx, `
		SELECT direction
		FROM wallet_ledger
		WHERE transaction_id = 'rollback-bet-test'
	`).Scan(&direction)

	if err != nil {
		t.Fatalf("failed to find rollback ledger: %v", err)
	}

	if direction != "CREDIT" {
		t.Fatalf("expected rollback ledger direction CREDIT, got %s", direction)
	}
}

func TestPostgresWagerTransactionProcessor_RollbackWin(t *testing.T) {
	ctx, pool, processor := setupPostgresWagerTransactionProcessorTest(t)

	now := time.Now()

	_, err := pool.Exec(ctx, `
		DELETE FROM wallet_ledger
		WHERE wallet_id = 'wallet-rollback-win-test'
	`)
	if err != nil {
		t.Fatalf("failed to clean ledger: %v", err)
	}

	_, err = pool.Exec(ctx, `
		DELETE FROM wager_transactions
		WHERE id IN ('win-rollback-test', 'rollback-win-test')
	`)
	if err != nil {
		t.Fatalf("failed to clean wager transactions: %v", err)
	}

	_, err = pool.Exec(ctx, `
		DELETE FROM wallets
		WHERE id = 'wallet-rollback-win-test'
	`)
	if err != nil {
		t.Fatalf("failed to clean wallet: %v", err)
	}

	wallet, err := domain.NewWallet(
		"wallet-rollback-win-test",
		"player-rollback-win-test",
		domain.NewMoney(10000, "BRL"),
		now,
	)
	if err != nil {
		t.Fatalf("failed to create wallet: %v", err)
	}

	walletRepository := NewPostgresWalletRepository(pool)

	if err := walletRepository.Create(ctx, wallet); err != nil {
		t.Fatalf("failed to persist wallet: %v", err)
	}

	win, err := domain.NewWagerTransaction(
		"win-rollback-test",
		"provider-rollback-win-test",
		"external-win-rollback-test",
		"idempotency-win-rollback-test",
		"hash-win-rollback-test",
		"player-rollback-win-test",
		"wallet-rollback-win-test",
		"round-rollback-win-test",
		"game-rollback-win-test",
		domain.TransactionWin,
		domain.NewMoney(3000, "BRL"),
		"",
		now,
	)
	if err != nil {
		t.Fatalf("failed to create win: %v", err)
	}

	if err := processor.Execute(ctx, win); err != nil {
		t.Fatalf("failed to process win: %v", err)
	}

	rollback, err := domain.NewWagerTransaction(
		"rollback-win-test",
		"provider-rollback-win-test",
		"external-rollback-win-test",
		"idempotency-rollback-win-test",
		"hash-rollback-win-test",
		"player-rollback-win-test",
		"wallet-rollback-win-test",
		"round-rollback-win-test",
		"game-rollback-win-test",
		domain.TransactionRollback,
		domain.NewMoney(3000, "BRL"),
		"external-win-rollback-test",
		now.Add(time.Second),
	)
	if err != nil {
		t.Fatalf("failed to create rollback: %v", err)
	}

	if err := processor.Execute(ctx, rollback); err != nil {
		t.Fatalf("failed to process rollback: %v", err)
	}

	result, err := walletRepository.FindByID(ctx, "wallet-rollback-win-test")
	if err != nil {
		t.Fatalf("failed to find wallet: %v", err)
	}

	if result.Balance.Amount != 10000 {
		t.Fatalf("expected final balance 10000, got %d", result.Balance.Amount)
	}

	if result.Version != 3 {
		t.Fatalf("expected wallet version 3, got %d", result.Version)
	}

	var direction string

	err = pool.QueryRow(ctx, `
		SELECT direction
		FROM wallet_ledger
		WHERE transaction_id = 'rollback-win-test'
	`).Scan(&direction)

	if err != nil {
		t.Fatalf("failed to find rollback ledger: %v", err)
	}

	if direction != "DEBIT" {
		t.Fatalf("expected rollback ledger direction DEBIT, got %s", direction)
	}
}

func TestPostgresWagerTransactionProcessor_RollbackRefund(t *testing.T) {
	ctx, pool, processor := setupPostgresWagerTransactionProcessorTest(t)

	now := time.Now()

	_, err := pool.Exec(ctx, `
		DELETE FROM wallet_ledger
		WHERE wallet_id = 'wallet-rollback-refund-test'
	`)
	if err != nil {
		t.Fatalf("failed to clean ledger: %v", err)
	}

	_, err = pool.Exec(ctx, `
		DELETE FROM wager_transactions
		WHERE id IN (
			'bet-rollback-refund-test',
			'refund-rollback-refund-test',
			'rollback-refund-test'
		)
	`)
	if err != nil {
		t.Fatalf("failed to clean wager transactions: %v", err)
	}

	_, err = pool.Exec(ctx, `
		DELETE FROM wallets
		WHERE id = 'wallet-rollback-refund-test'
	`)
	if err != nil {
		t.Fatalf("failed to clean wallet: %v", err)
	}

	wallet, err := domain.NewWallet(
		"wallet-rollback-refund-test",
		"player-rollback-refund-test",
		domain.NewMoney(10000, "BRL"),
		now,
	)
	if err != nil {
		t.Fatalf("failed to create wallet: %v", err)
	}

	walletRepository := NewPostgresWalletRepository(pool)

	if err := walletRepository.Create(ctx, wallet); err != nil {
		t.Fatalf("failed to persist wallet: %v", err)
	}

	bet, err := domain.NewWagerTransaction(
		"bet-rollback-refund-test",
		"provider-rollback-refund-test",
		"external-bet-rollback-refund-test",
		"idempotency-bet-rollback-refund-test",
		"hash-bet-rollback-refund-test",
		"player-rollback-refund-test",
		"wallet-rollback-refund-test",
		"round-rollback-refund-test",
		"game-rollback-refund-test",
		domain.TransactionBet,
		domain.NewMoney(3000, "BRL"),
		"",
		now,
	)
	if err != nil {
		t.Fatalf("failed to create bet: %v", err)
	}

	if err := processor.Execute(ctx, bet); err != nil {
		t.Fatalf("failed to process bet: %v", err)
	}

	refund, err := domain.NewWagerTransaction(
		"refund-rollback-refund-test",
		"provider-rollback-refund-test",
		"external-refund-rollback-refund-test",
		"idempotency-refund-rollback-refund-test",
		"hash-refund-rollback-refund-test",
		"player-rollback-refund-test",
		"wallet-rollback-refund-test",
		"round-rollback-refund-test",
		"game-rollback-refund-test",
		domain.TransactionRefund,
		domain.NewMoney(3000, "BRL"),
		"external-bet-rollback-refund-test",
		now.Add(time.Second),
	)
	if err != nil {
		t.Fatalf("failed to create refund: %v", err)
	}

	if err := processor.Execute(ctx, refund); err != nil {
		t.Fatalf("failed to process refund: %v", err)
	}

	rollback, err := domain.NewWagerTransaction(
		"rollback-refund-test",
		"provider-rollback-refund-test",
		"external-rollback-refund-test",
		"idempotency-rollback-refund-test",
		"hash-rollback-refund-test",
		"player-rollback-refund-test",
		"wallet-rollback-refund-test",
		"round-rollback-refund-test",
		"game-rollback-refund-test",
		domain.TransactionRollback,
		domain.NewMoney(3000, "BRL"),
		"external-refund-rollback-refund-test",
		now.Add(2*time.Second),
	)
	if err != nil {
		t.Fatalf("failed to create rollback: %v", err)
	}

	if err := processor.Execute(ctx, rollback); err != nil {
		t.Fatalf("failed to process rollback: %v", err)
	}

	result, err := walletRepository.FindByID(ctx, "wallet-rollback-refund-test")
	if err != nil {
		t.Fatalf("failed to find wallet: %v", err)
	}

	if result.Balance.Amount != 7000 {
		t.Fatalf("expected final balance 7000, got %d", result.Balance.Amount)
	}

	if result.Version != 4 {
		t.Fatalf("expected wallet version 4, got %d", result.Version)
	}

	var direction string

	err = pool.QueryRow(ctx, `
		SELECT direction
		FROM wallet_ledger
		WHERE transaction_id = 'rollback-refund-test'
	`).Scan(&direction)

	if err != nil {
		t.Fatalf("failed to find rollback ledger: %v", err)
	}

	if direction != "DEBIT" {
		t.Fatalf("expected rollback ledger direction DEBIT, got %s", direction)
	}
}

func TestPostgresWagerTransactionProcessor_RejectsRefundOfWin(t *testing.T) {
	ctx, pool, processor := setupPostgresWagerTransactionProcessorTest(t)

	now := time.Now()

	_, err := pool.Exec(ctx, `
		DELETE FROM wallet_ledger
		WHERE wallet_id = 'wallet-invalid-refund-test'
	`)
	if err != nil {
		t.Fatalf("failed to clean ledger: %v", err)
	}

	_, err = pool.Exec(ctx, `
		DELETE FROM wager_transactions
		WHERE id IN ('win-invalid-refund-test', 'refund-invalid-refund-test')
	`)
	if err != nil {
		t.Fatalf("failed to clean wager transactions: %v", err)
	}

	_, err = pool.Exec(ctx, `
		DELETE FROM wallets
		WHERE id = 'wallet-invalid-refund-test'
	`)
	if err != nil {
		t.Fatalf("failed to clean wallet: %v", err)
	}

	wallet, err := domain.NewWallet(
		"wallet-invalid-refund-test",
		"player-invalid-refund-test",
		domain.NewMoney(10000, "BRL"),
		now,
	)
	if err != nil {
		t.Fatalf("failed to create wallet: %v", err)
	}

	walletRepository := NewPostgresWalletRepository(pool)

	if err := walletRepository.Create(ctx, wallet); err != nil {
		t.Fatalf("failed to persist wallet: %v", err)
	}

	win, err := domain.NewWagerTransaction(
		"win-invalid-refund-test",
		"provider-invalid-refund-test",
		"external-win-invalid-refund-test",
		"idempotency-win-invalid-refund-test",
		"hash-win-invalid-refund-test",
		"player-invalid-refund-test",
		"wallet-invalid-refund-test",
		"round-invalid-refund-test",
		"game-invalid-refund-test",
		domain.TransactionWin,
		domain.NewMoney(3000, "BRL"),
		"",
		now,
	)
	if err != nil {
		t.Fatalf("failed to create win: %v", err)
	}

	if err := processor.Execute(ctx, win); err != nil {
		t.Fatalf("failed to process win: %v", err)
	}

	refund, err := domain.NewWagerTransaction(
		"refund-invalid-refund-test",
		"provider-invalid-refund-test",
		"external-refund-invalid-refund-test",
		"idempotency-refund-invalid-refund-test",
		"hash-refund-invalid-refund-test",
		"player-invalid-refund-test",
		"wallet-invalid-refund-test",
		"round-invalid-refund-test",
		"game-invalid-refund-test",
		domain.TransactionRefund,
		domain.NewMoney(3000, "BRL"),
		"external-win-invalid-refund-test",
		now.Add(time.Second),
	)
	if err != nil {
		t.Fatalf("failed to create refund: %v", err)
	}

	err = processor.Execute(ctx, refund)

	if !errors.Is(err, ErrInvalidReference) {
		t.Fatalf("expected ErrInvalidReference, got %v", err)
	}

	result, err := walletRepository.FindByID(ctx, "wallet-invalid-refund-test")
	if err != nil {
		t.Fatalf("failed to find wallet: %v", err)
	}

	if result.Balance.Amount != 13000 {
		t.Fatalf("expected balance to remain 13000, got %d", result.Balance.Amount)
	}

	var state domain.TransactionState
	var failureCode string

	err = pool.QueryRow(ctx, `
	SELECT state, failure_code
	FROM wager_transactions
	WHERE id = 'refund-invalid-refund-test'
`).Scan(&state, &failureCode)

	if err != nil {
		t.Fatalf("failed to check refund transaction: %v", err)
	}

	if state != domain.TransactionRejected {
		t.Fatalf("expected refund transaction to be rejected, got %s", state)
	}

	if failureCode != FailureCodeInvalidReference {
		t.Fatalf(
			"expected failure code %s, got %s",
			FailureCodeInvalidReference,
			failureCode,
		)
	}

}

func TestPostgresWagerTransactionProcessor_RejectsMissingRefundReference(t *testing.T) {
	ctx, pool, processor := setupPostgresWagerTransactionProcessorTest(t)

	now := time.Now()

	_, err := pool.Exec(ctx, `
		DELETE FROM wallet_ledger
		WHERE wallet_id = 'wallet-missing-reference-test'
	`)
	if err != nil {
		t.Fatalf("failed to clean ledger: %v", err)
	}

	_, err = pool.Exec(ctx, `
		DELETE FROM wager_transactions
		WHERE id = 'refund-missing-reference-test'
	`)
	if err != nil {
		t.Fatalf("failed to clean wager transactions: %v", err)
	}

	_, err = pool.Exec(ctx, `
		DELETE FROM wallets
		WHERE id = 'wallet-missing-reference-test'
	`)
	if err != nil {
		t.Fatalf("failed to clean wallet: %v", err)
	}

	wallet, err := domain.NewWallet(
		"wallet-missing-reference-test",
		"player-missing-reference-test",
		domain.NewMoney(10000, "BRL"),
		now,
	)
	if err != nil {
		t.Fatalf("failed to create wallet: %v", err)
	}

	walletRepository := NewPostgresWalletRepository(pool)

	if err := walletRepository.Create(ctx, wallet); err != nil {
		t.Fatalf("failed to persist wallet: %v", err)
	}

	refund, err := domain.NewWagerTransaction(
		"refund-missing-reference-test",
		"provider-missing-reference-test",
		"external-refund-missing-reference-test",
		"idempotency-refund-missing-reference-test",
		"hash-refund-missing-reference-test",
		"player-missing-reference-test",
		"wallet-missing-reference-test",
		"round-missing-reference-test",
		"game-missing-reference-test",
		domain.TransactionRefund,
		domain.NewMoney(3000, "BRL"),
		"external-bet-does-not-exist",
		now,
	)
	if err != nil {
		t.Fatalf("failed to create refund: %v", err)
	}

	err = processor.Execute(ctx, refund)

	if err != nil {
		t.Fatalf("expected refund to remain pending, got %v", err)
	}

	result, err := walletRepository.FindByID(
		ctx,
		"wallet-missing-reference-test",
	)
	if err != nil {
		t.Fatalf("failed to find wallet: %v", err)
	}

	if result.Balance.Amount != 10000 {
		t.Fatalf("expected balance to remain 10000, got %d", result.Balance.Amount)
	}

	if result.Version != 1 {
		t.Fatalf("expected wallet version 1, got %d", result.Version)
	}

	var state string

	err = pool.QueryRow(ctx, `
		SELECT state
		FROM wager_transactions
		WHERE id = 'refund-missing-reference-test'
	`).Scan(&state)

	if err != nil {
		t.Fatalf("failed to find refund transaction: %v", err)
	}

	if state != string(domain.TransactionPendingReference) {
		t.Fatalf(
			"expected refund transaction to be PENDING_REFERENCE, got %s",
			state,
		)
	}

	var ledgerCount int

	err = pool.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM wallet_ledger
		WHERE wallet_id = 'wallet-missing-reference-test'
	`).Scan(&ledgerCount)

	if err != nil {
		t.Fatalf("failed to count ledger entries: %v", err)
	}

	if ledgerCount != 0 {
		t.Fatalf(
			"expected no ledger entries for pending refund, got %d",
			ledgerCount,
		)
	}
}

func TestPostgresWagerTransactionProcessor_MissingRollbackReference(t *testing.T) {
	ctx, pool, processor := setupPostgresWagerTransactionProcessorTest(t)

	now := time.Now()

	_, err := pool.Exec(ctx, `
		DELETE FROM wallet_ledger
		WHERE wallet_id = 'wallet-missing-rollback-reference-test'
	`)
	if err != nil {
		t.Fatalf("failed to clean ledger: %v", err)
	}

	_, err = pool.Exec(ctx, `
		DELETE FROM wager_transactions
		WHERE id = 'rollback-missing-reference-test'
	`)
	if err != nil {
		t.Fatalf("failed to clean wager transactions: %v", err)
	}

	_, err = pool.Exec(ctx, `
		DELETE FROM wallets
		WHERE id = 'wallet-missing-rollback-reference-test'
	`)
	if err != nil {
		t.Fatalf("failed to clean wallet: %v", err)
	}

	wallet, err := domain.NewWallet(
		"wallet-missing-rollback-reference-test",
		"player-missing-rollback-reference-test",
		domain.NewMoney(10000, "BRL"),
		now,
	)
	if err != nil {
		t.Fatalf("failed to create wallet: %v", err)
	}

	walletRepository := NewPostgresWalletRepository(pool)

	if err := walletRepository.Create(ctx, wallet); err != nil {
		t.Fatalf("failed to persist wallet: %v", err)
	}

	rollback, err := domain.NewWagerTransaction(
		"rollback-missing-reference-test",
		"provider-missing-rollback-reference-test",
		"external-rollback-missing-reference-test",
		"idempotency-rollback-missing-reference-test",
		"hash-rollback-missing-reference-test",
		"player-missing-rollback-reference-test",
		"wallet-missing-rollback-reference-test",
		"round-missing-rollback-reference-test",
		"game-missing-rollback-reference-test",
		domain.TransactionRollback,
		domain.NewMoney(3000, "BRL"),
		"external-bet-does-not-exist",
		now,
	)
	if err != nil {
		t.Fatalf("failed to create rollback: %v", err)
	}

	err = processor.Execute(ctx, rollback)

	if err != nil {
		t.Fatalf("expected rollback to remain pending, got %v", err)
	}

	result, err := walletRepository.FindByID(
		ctx,
		"wallet-missing-rollback-reference-test",
	)
	if err != nil {
		t.Fatalf("failed to find wallet: %v", err)
	}

	if result.Balance.Amount != 10000 {
		t.Fatalf(
			"expected balance to remain 10000, got %d",
			result.Balance.Amount,
		)
	}

	if result.Version != 1 {
		t.Fatalf("expected wallet version 1, got %d", result.Version)
	}

	var state string

	err = pool.QueryRow(ctx, `
		SELECT state
		FROM wager_transactions
		WHERE id = 'rollback-missing-reference-test'
	`).Scan(&state)

	if err != nil {
		t.Fatalf("failed to find rollback transaction: %v", err)
	}

	if state != string(domain.TransactionPendingReference) {
		t.Fatalf(
			"expected rollback transaction to be PENDING_REFERENCE, got %s",
			state,
		)
	}

	var ledgerCount int

	err = pool.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM wallet_ledger
		WHERE wallet_id = 'wallet-missing-rollback-reference-test'
	`).Scan(&ledgerCount)

	if err != nil {
		t.Fatalf("failed to count ledger entries: %v", err)
	}

	if ledgerCount != 0 {
		t.Fatalf(
			"expected no ledger entries for pending rollback, got %d",
			ledgerCount,
		)
	}
}

func TestPostgresWagerTransactionProcessor_RejectsDifferentRefundAmount(t *testing.T) {
	ctx, pool, processor := setupPostgresWagerTransactionProcessorTest(t)

	now := time.Now()

	_, err := pool.Exec(ctx, `
		DELETE FROM wallet_ledger
		WHERE wallet_id = 'wallet-different-refund-test'
	`)
	if err != nil {
		t.Fatalf("failed to clean ledger: %v", err)
	}

	_, err = pool.Exec(ctx, `
		DELETE FROM wager_transactions
		WHERE id IN ('bet-different-refund-test', 'refund-different-refund-test')
	`)
	if err != nil {
		t.Fatalf("failed to clean wager transactions: %v", err)
	}

	_, err = pool.Exec(ctx, `
		DELETE FROM wallets
		WHERE id = 'wallet-different-refund-test'
	`)
	if err != nil {
		t.Fatalf("failed to clean wallet: %v", err)
	}

	wallet, err := domain.NewWallet(
		"wallet-different-refund-test",
		"player-different-refund-test",
		domain.NewMoney(10000, "BRL"),
		now,
	)
	if err != nil {
		t.Fatalf("failed to create wallet: %v", err)
	}

	walletRepository := NewPostgresWalletRepository(pool)

	if err := walletRepository.Create(ctx, wallet); err != nil {
		t.Fatalf("failed to persist wallet: %v", err)
	}

	bet, err := domain.NewWagerTransaction(
		"bet-different-refund-test",
		"provider-different-refund-test",
		"external-bet-different-refund-test",
		"idempotency-bet-different-refund-test",
		"hash-bet-different-refund-test",
		"player-different-refund-test",
		"wallet-different-refund-test",
		"round-different-refund-test",
		"game-different-refund-test",
		domain.TransactionBet,
		domain.NewMoney(3000, "BRL"),
		"",
		now,
	)
	if err != nil {
		t.Fatalf("failed to create bet: %v", err)
	}

	if err := processor.Execute(ctx, bet); err != nil {
		t.Fatalf("failed to process bet: %v", err)
	}

	refund, err := domain.NewWagerTransaction(
		"refund-different-refund-test",
		"provider-different-refund-test",
		"external-refund-different-refund-test",
		"idempotency-refund-different-refund-test",
		"hash-refund-different-refund-test",
		"player-different-refund-test",
		"wallet-different-refund-test",
		"round-different-refund-test",
		"game-different-refund-test",
		domain.TransactionRefund,
		domain.NewMoney(2000, "BRL"),
		"external-bet-different-refund-test",
		now.Add(time.Second),
	)
	if err != nil {
		t.Fatalf("failed to create refund: %v", err)
	}

	err = processor.Execute(ctx, refund)

	if !errors.Is(err, ErrReferenceMismatch) {
		t.Fatalf("expected ErrReferenceMismatch, got %v", err)
	}

	result, err := walletRepository.FindByID(ctx, "wallet-different-refund-test")
	if err != nil {
		t.Fatalf("failed to find wallet: %v", err)
	}

	if result.Balance.Amount != 7000 {
		t.Fatalf("expected balance to remain 7000, got %d", result.Balance.Amount)
	}

	if result.Version != 2 {
		t.Fatalf("expected wallet version 2, got %d", result.Version)
	}

	var state domain.TransactionState
	var failureCode string

	err = pool.QueryRow(ctx, `
	SELECT state, failure_code
	FROM wager_transactions
	WHERE id = 'refund-different-refund-test'
`).Scan(&state, &failureCode)

	if err != nil {
		t.Fatalf("failed to check refund transaction: %v", err)
	}

	if state != domain.TransactionRejected {
		t.Fatalf(
			"expected refund transaction to be rejected, got %s",
			state,
		)
	}

	if failureCode != FailureCodeReferenceMismatch {
		t.Fatalf(
			"expected failure code %s, got %s",
			FailureCodeReferenceMismatch,
			failureCode,
		)
	}
}

func TestPostgresWagerTransactionProcessor_RejectsDuplicateRefund(t *testing.T) {
	ctx, pool, processor := setupPostgresWagerTransactionProcessorTest(t)

	now := time.Now()

	_, err := pool.Exec(ctx, `
		DELETE FROM wallet_ledger
		WHERE wallet_id = 'wallet-duplicate-refund-test'
	`)
	if err != nil {
		t.Fatalf("failed to clean ledger: %v", err)
	}

	_, err = pool.Exec(ctx, `
		DELETE FROM wager_transactions
		WHERE id IN (
			'bet-duplicate-refund-test',
			'refund-duplicate-refund-1-test',
			'refund-duplicate-refund-2-test'
		)
	`)
	if err != nil {
		t.Fatalf("failed to clean wager transactions: %v", err)
	}

	_, err = pool.Exec(ctx, `
		DELETE FROM wallets
		WHERE id = 'wallet-duplicate-refund-test'
	`)
	if err != nil {
		t.Fatalf("failed to clean wallet: %v", err)
	}

	wallet, err := domain.NewWallet(
		"wallet-duplicate-refund-test",
		"player-duplicate-refund-test",
		domain.NewMoney(10000, "BRL"),
		now,
	)
	if err != nil {
		t.Fatalf("failed to create wallet: %v", err)
	}

	walletRepository := NewPostgresWalletRepository(pool)

	if err := walletRepository.Create(ctx, wallet); err != nil {
		t.Fatalf("failed to persist wallet: %v", err)
	}

	bet, err := domain.NewWagerTransaction(
		"bet-duplicate-refund-test",
		"provider-duplicate-refund-test",
		"external-bet-duplicate-refund-test",
		"idempotency-bet-duplicate-refund-test",
		"hash-bet-duplicate-refund-test",
		"player-duplicate-refund-test",
		"wallet-duplicate-refund-test",
		"round-duplicate-refund-test",
		"game-duplicate-refund-test",
		domain.TransactionBet,
		domain.NewMoney(3000, "BRL"),
		"",
		now,
	)
	if err != nil {
		t.Fatalf("failed to create bet: %v", err)
	}

	if err := processor.Execute(ctx, bet); err != nil {
		t.Fatalf("failed to process bet: %v", err)
	}

	firstRefund, err := domain.NewWagerTransaction(
		"refund-duplicate-refund-1-test",
		"provider-duplicate-refund-test",
		"external-refund-duplicate-1-test",
		"idempotency-refund-duplicate-1-test",
		"hash-refund-duplicate-1-test",
		"player-duplicate-refund-test",
		"wallet-duplicate-refund-test",
		"round-duplicate-refund-test",
		"game-duplicate-refund-test",
		domain.TransactionRefund,
		domain.NewMoney(3000, "BRL"),
		"external-bet-duplicate-refund-test",
		now.Add(time.Second),
	)
	if err != nil {
		t.Fatalf("failed to create first refund: %v", err)
	}

	if err := processor.Execute(ctx, firstRefund); err != nil {
		t.Fatalf("failed to process first refund: %v", err)
	}

	secondRefund, err := domain.NewWagerTransaction(
		"refund-duplicate-refund-2-test",
		"provider-duplicate-refund-test",
		"external-refund-duplicate-2-test",
		"idempotency-refund-duplicate-2-test",
		"hash-refund-duplicate-2-test",
		"player-duplicate-refund-test",
		"wallet-duplicate-refund-test",
		"round-duplicate-refund-test",
		"game-duplicate-refund-test",
		domain.TransactionRefund,
		domain.NewMoney(3000, "BRL"),
		"external-bet-duplicate-refund-test",
		now.Add(2*time.Second),
	)
	if err != nil {
		t.Fatalf("failed to create second refund: %v", err)
	}

	err = processor.Execute(ctx, secondRefund)

	if !errors.Is(err, ErrDuplicateReversal) {
		t.Fatalf("expected ErrDuplicateReversal, got %v", err)
	}

	result, err := walletRepository.FindByID(ctx, "wallet-duplicate-refund-test")
	if err != nil {
		t.Fatalf("failed to find wallet: %v", err)
	}

	if result.Balance.Amount != 10000 {
		t.Fatalf("expected balance to remain 10000, got %d", result.Balance.Amount)
	}

	if result.Version != 3 {
		t.Fatalf("expected wallet version 3, got %d", result.Version)
	}

	var count int

	err = pool.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM wager_transactions
		WHERE reference_external_transaction_id = 'external-bet-duplicate-refund-test'
		  AND transaction_type = 'REFUND'
		  AND state = 'PROCESSED'
	`).Scan(&count)

	if err != nil {
		t.Fatalf("failed to count processed refunds: %v", err)
	}

	if count != 1 {
		t.Fatalf("expected exactly 1 processed refund, got %d", count)
	}
}

func TestPostgresWagerTransactionProcessor_ConcurrentDuplicateRefund(t *testing.T) {
	ctx, pool, processor := setupPostgresWagerTransactionProcessorTest(t)

	now := time.Now()

	_, err := pool.Exec(ctx, `
		DELETE FROM wallet_ledger
		WHERE wallet_id = 'wallet-concurrent-refund-test'
	`)
	if err != nil {
		t.Fatalf("failed to clean ledger: %v", err)
	}

	_, err = pool.Exec(ctx, `
		DELETE FROM wager_transactions
		WHERE id IN (
			'bet-concurrent-refund-test',
			'refund-concurrent-refund-1-test',
			'refund-concurrent-refund-2-test'
		)
	`)
	if err != nil {
		t.Fatalf("failed to clean wager transactions: %v", err)
	}

	_, err = pool.Exec(ctx, `
		DELETE FROM wallets
		WHERE id = 'wallet-concurrent-refund-test'
	`)
	if err != nil {
		t.Fatalf("failed to clean wallet: %v", err)
	}

	wallet, err := domain.NewWallet(
		"wallet-concurrent-refund-test",
		"player-concurrent-refund-test",
		domain.NewMoney(10000, "BRL"),
		now,
	)
	if err != nil {
		t.Fatalf("failed to create wallet: %v", err)
	}

	walletRepository := NewPostgresWalletRepository(pool)

	if err := walletRepository.Create(ctx, wallet); err != nil {
		t.Fatalf("failed to persist wallet: %v", err)
	}

	bet, err := domain.NewWagerTransaction(
		"bet-concurrent-refund-test",
		"provider-concurrent-refund-test",
		"external-bet-concurrent-refund-test",
		"idempotency-bet-concurrent-refund-test",
		"hash-bet-concurrent-refund-test",
		"player-concurrent-refund-test",
		"wallet-concurrent-refund-test",
		"round-concurrent-refund-test",
		"game-concurrent-refund-test",
		domain.TransactionBet,
		domain.NewMoney(3000, "BRL"),
		"",
		now,
	)
	if err != nil {
		t.Fatalf("failed to create bet: %v", err)
	}

	if err := processor.Execute(ctx, bet); err != nil {
		t.Fatalf("failed to process bet: %v", err)
	}

	firstRefund, err := domain.NewWagerTransaction(
		"refund-concurrent-refund-1-test",
		"provider-concurrent-refund-test",
		"external-refund-concurrent-1-test",
		"idempotency-refund-concurrent-1-test",
		"hash-refund-concurrent-1-test",
		"player-concurrent-refund-test",
		"wallet-concurrent-refund-test",
		"round-concurrent-refund-test",
		"game-concurrent-refund-test",
		domain.TransactionRefund,
		domain.NewMoney(3000, "BRL"),
		"external-bet-concurrent-refund-test",
		now.Add(time.Second),
	)
	if err != nil {
		t.Fatalf("failed to create first refund: %v", err)
	}

	secondRefund, err := domain.NewWagerTransaction(
		"refund-concurrent-refund-2-test",
		"provider-concurrent-refund-test",
		"external-refund-concurrent-2-test",
		"idempotency-refund-concurrent-2-test",
		"hash-refund-concurrent-2-test",
		"player-concurrent-refund-test",
		"wallet-concurrent-refund-test",
		"round-concurrent-refund-test",
		"game-concurrent-refund-test",
		domain.TransactionRefund,
		domain.NewMoney(3000, "BRL"),
		"external-bet-concurrent-refund-test",
		now.Add(2*time.Second),
	)
	if err != nil {
		t.Fatalf("failed to create second refund: %v", err)
	}

	start := make(chan struct{})

	var wg sync.WaitGroup
	wg.Add(2)

	errs := make(chan error, 2)

	go func() {
		defer wg.Done()

		<-start

		errs <- processor.Execute(ctx, firstRefund)
	}()

	go func() {
		defer wg.Done()

		<-start

		errs <- processor.Execute(ctx, secondRefund)
	}()

	close(start)
	wg.Wait()
	close(errs)

	var processedCount int
	var rejectedCount int

	rows, err := pool.Query(ctx, `
		SELECT state
		FROM wager_transactions
		WHERE id IN (
			'refund-concurrent-refund-1-test',
			'refund-concurrent-refund-2-test'
		)
	`)
	if err != nil {
		t.Fatalf("failed to query refunds: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var state string

		if err := rows.Scan(&state); err != nil {
			t.Fatalf("failed to scan refund state: %v", err)
		}

		switch state {
		case string(domain.TransactionProcessed):
			processedCount++
		case string(domain.TransactionRejected):
			rejectedCount++
		}
	}

	if err := rows.Err(); err != nil {
		t.Fatalf("failed to iterate refunds: %v", err)
	}

	if processedCount != 1 {
		t.Fatalf("expected exactly 1 processed refund, got %d", processedCount)
	}

	if rejectedCount != 1 {
		t.Fatalf("expected exactly 1 rejected refund, got %d", rejectedCount)
	}

	var duplicateError bool

	for err := range errs {
		if errors.Is(err, ErrDuplicateReversal) {
			duplicateError = true
		}
	}

	if !duplicateError {
		t.Fatal("expected one refund to fail with ErrDuplicateReversal")
	}

	result, err := walletRepository.FindByID(ctx, "wallet-concurrent-refund-test")
	if err != nil {
		t.Fatalf("failed to find wallet: %v", err)
	}

	if result.Balance.Amount != 10000 {
		t.Fatalf("expected balance to be 10000, got %d", result.Balance.Amount)
	}

	if result.Version != 3 {
		t.Fatalf("expected wallet version 3, got %d", result.Version)
	}

	var ledgerCount int

	err = pool.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM wallet_ledger
		WHERE wallet_id = 'wallet-concurrent-refund-test'
	`).Scan(&ledgerCount)

	if err != nil {
		t.Fatalf("failed to count ledger entries: %v", err)
	}

	if ledgerCount != 2 {
		t.Fatalf("expected exactly 2 ledger entries, got %d", ledgerCount)
	}
}

func TestPostgresWagerTransactionProcessor_RejectsDuplicateRollback(t *testing.T) {
	ctx, pool, processor := setupPostgresWagerTransactionProcessorTest(t)

	now := time.Now()

	_, err := pool.Exec(ctx, `
		DELETE FROM wallet_ledger
		WHERE wallet_id = 'wallet-duplicate-rollback-test'
	`)
	if err != nil {
		t.Fatalf("failed to clean ledger: %v", err)
	}

	_, err = pool.Exec(ctx, `
		DELETE FROM wager_transactions
		WHERE id IN (
			'bet-duplicate-rollback-test',
			'rollback-duplicate-rollback-1-test',
			'rollback-duplicate-rollback-2-test'
		)
	`)
	if err != nil {
		t.Fatalf("failed to clean wager transactions: %v", err)
	}

	_, err = pool.Exec(ctx, `
		DELETE FROM wallets
		WHERE id = 'wallet-duplicate-rollback-test'
	`)
	if err != nil {
		t.Fatalf("failed to clean wallet: %v", err)
	}

	wallet, err := domain.NewWallet(
		"wallet-duplicate-rollback-test",
		"player-duplicate-rollback-test",
		domain.NewMoney(10000, "BRL"),
		now,
	)
	if err != nil {
		t.Fatalf("failed to create wallet: %v", err)
	}

	walletRepository := NewPostgresWalletRepository(pool)

	if err := walletRepository.Create(ctx, wallet); err != nil {
		t.Fatalf("failed to persist wallet: %v", err)
	}

	bet, err := domain.NewWagerTransaction(
		"bet-duplicate-rollback-test",
		"provider-duplicate-rollback-test",
		"external-bet-duplicate-rollback-test",
		"idempotency-bet-duplicate-rollback-test",
		"hash-bet-duplicate-rollback-test",
		"player-duplicate-rollback-test",
		"wallet-duplicate-rollback-test",
		"round-duplicate-rollback-test",
		"game-duplicate-rollback-test",
		domain.TransactionBet,
		domain.NewMoney(3000, "BRL"),
		"",
		now,
	)
	if err != nil {
		t.Fatalf("failed to create bet: %v", err)
	}

	if err := processor.Execute(ctx, bet); err != nil {
		t.Fatalf("failed to process bet: %v", err)
	}

	firstRollback, err := domain.NewWagerTransaction(
		"rollback-duplicate-rollback-1-test",
		"provider-duplicate-rollback-test",
		"external-rollback-duplicate-1-test",
		"idempotency-rollback-duplicate-1-test",
		"hash-rollback-duplicate-1-test",
		"player-duplicate-rollback-test",
		"wallet-duplicate-rollback-test",
		"round-duplicate-rollback-test",
		"game-duplicate-rollback-test",
		domain.TransactionRollback,
		domain.NewMoney(3000, "BRL"),
		"external-bet-duplicate-rollback-test",
		now.Add(time.Second),
	)
	if err != nil {
		t.Fatalf("failed to create first rollback: %v", err)
	}

	if err := processor.Execute(ctx, firstRollback); err != nil {
		t.Fatalf("failed to process first rollback: %v", err)
	}

	secondRollback, err := domain.NewWagerTransaction(
		"rollback-duplicate-rollback-2-test",
		"provider-duplicate-rollback-test",
		"external-rollback-duplicate-2-test",
		"idempotency-rollback-duplicate-2-test",
		"hash-rollback-duplicate-2-test",
		"player-duplicate-rollback-test",
		"wallet-duplicate-rollback-test",
		"round-duplicate-rollback-test",
		"game-duplicate-rollback-test",
		domain.TransactionRollback,
		domain.NewMoney(3000, "BRL"),
		"external-bet-duplicate-rollback-test",
		now.Add(2*time.Second),
	)
	if err != nil {
		t.Fatalf("failed to create second rollback: %v", err)
	}

	err = processor.Execute(ctx, secondRollback)

	if !errors.Is(err, ErrDuplicateReversal) {
		t.Fatalf("expected ErrDuplicateReversal, got %v", err)
	}

	result, err := walletRepository.FindByID(ctx, "wallet-duplicate-rollback-test")
	if err != nil {
		t.Fatalf("failed to find wallet: %v", err)
	}

	if result.Balance.Amount != 10000 {
		t.Fatalf("expected balance to remain 10000, got %d", result.Balance.Amount)
	}

	if result.Version != 3 {
		t.Fatalf("expected wallet version 3, got %d", result.Version)
	}

	var count int

	err = pool.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM wager_transactions
		WHERE reference_external_transaction_id = 'external-bet-duplicate-rollback-test'
		  AND transaction_type = 'ROLLBACK'
		  AND state = 'PROCESSED'
	`).Scan(&count)

	if err != nil {
		t.Fatalf("failed to count processed rollbacks: %v", err)
	}

	if count != 1 {
		t.Fatalf("expected exactly 1 processed rollback, got %d", count)
	}
}

func TestPostgresWagerTransactionProcessor_ConcurrentDuplicateRollback(t *testing.T) {
	ctx, pool, processor := setupPostgresWagerTransactionProcessorTest(t)

	now := time.Now()

	_, err := pool.Exec(ctx, `
		DELETE FROM wallet_ledger
		WHERE wallet_id = 'wallet-concurrent-rollback-test'
	`)
	if err != nil {
		t.Fatalf("failed to clean ledger: %v", err)
	}

	_, err = pool.Exec(ctx, `
		DELETE FROM wager_transactions
		WHERE id IN (
			'bet-concurrent-rollback-test',
			'rollback-concurrent-rollback-1-test',
			'rollback-concurrent-rollback-2-test'
		)
	`)
	if err != nil {
		t.Fatalf("failed to clean wager transactions: %v", err)
	}

	_, err = pool.Exec(ctx, `
		DELETE FROM wallets
		WHERE id = 'wallet-concurrent-rollback-test'
	`)
	if err != nil {
		t.Fatalf("failed to clean wallet: %v", err)
	}

	wallet, err := domain.NewWallet(
		"wallet-concurrent-rollback-test",
		"player-concurrent-rollback-test",
		domain.NewMoney(10000, "BRL"),
		now,
	)
	if err != nil {
		t.Fatalf("failed to create wallet: %v", err)
	}

	walletRepository := NewPostgresWalletRepository(pool)

	if err := walletRepository.Create(ctx, wallet); err != nil {
		t.Fatalf("failed to persist wallet: %v", err)
	}

	bet, err := domain.NewWagerTransaction(
		"bet-concurrent-rollback-test",
		"provider-concurrent-rollback-test",
		"external-bet-concurrent-rollback-test",
		"idempotency-bet-concurrent-rollback-test",
		"hash-bet-concurrent-rollback-test",
		"player-concurrent-rollback-test",
		"wallet-concurrent-rollback-test",
		"round-concurrent-rollback-test",
		"game-concurrent-rollback-test",
		domain.TransactionBet,
		domain.NewMoney(3000, "BRL"),
		"",
		now,
	)
	if err != nil {
		t.Fatalf("failed to create bet: %v", err)
	}

	if err := processor.Execute(ctx, bet); err != nil {
		t.Fatalf("failed to process bet: %v", err)
	}

	firstRollback, err := domain.NewWagerTransaction(
		"rollback-concurrent-rollback-1-test",
		"provider-concurrent-rollback-test",
		"external-rollback-concurrent-1-test",
		"idempotency-rollback-concurrent-1-test",
		"hash-rollback-concurrent-1-test",
		"player-concurrent-rollback-test",
		"wallet-concurrent-rollback-test",
		"round-concurrent-rollback-test",
		"game-concurrent-rollback-test",
		domain.TransactionRollback,
		domain.NewMoney(3000, "BRL"),
		"external-bet-concurrent-rollback-test",
		now.Add(time.Second),
	)
	if err != nil {
		t.Fatalf("failed to create first rollback: %v", err)
	}

	secondRollback, err := domain.NewWagerTransaction(
		"rollback-concurrent-rollback-2-test",
		"provider-concurrent-rollback-test",
		"external-rollback-concurrent-2-test",
		"idempotency-rollback-concurrent-2-test",
		"hash-rollback-concurrent-2-test",
		"player-concurrent-rollback-test",
		"wallet-concurrent-rollback-test",
		"round-concurrent-rollback-test",
		"game-concurrent-rollback-test",
		domain.TransactionRollback,
		domain.NewMoney(3000, "BRL"),
		"external-bet-concurrent-rollback-test",
		now.Add(2*time.Second),
	)
	if err != nil {
		t.Fatalf("failed to create second rollback: %v", err)
	}

	start := make(chan struct{})

	var wg sync.WaitGroup
	wg.Add(2)

	errs := make(chan error, 2)

	go func() {
		defer wg.Done()

		<-start

		errs <- processor.Execute(ctx, firstRollback)
	}()

	go func() {
		defer wg.Done()

		<-start

		errs <- processor.Execute(ctx, secondRollback)
	}()

	close(start)
	wg.Wait()
	close(errs)

	var processedCount int
	var rejectedCount int

	rows, err := pool.Query(ctx, `
		SELECT state
		FROM wager_transactions
		WHERE id IN (
			'rollback-concurrent-rollback-1-test',
			'rollback-concurrent-rollback-2-test'
		)
	`)
	if err != nil {
		t.Fatalf("failed to query rollbacks: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var state string

		if err := rows.Scan(&state); err != nil {
			t.Fatalf("failed to scan rollback state: %v", err)
		}

		switch state {
		case string(domain.TransactionProcessed):
			processedCount++
		case string(domain.TransactionRejected):
			rejectedCount++
		}
	}

	if err := rows.Err(); err != nil {
		t.Fatalf("failed to iterate rollbacks: %v", err)
	}

	if processedCount != 1 {
		t.Fatalf("expected exactly 1 processed rollback, got %d", processedCount)
	}

	if rejectedCount != 1 {
		t.Fatalf("expected exactly 1 rejected rollback, got %d", rejectedCount)
	}

	var duplicateError bool

	for err := range errs {
		if errors.Is(err, ErrDuplicateReversal) {
			duplicateError = true
		}
	}

	if !duplicateError {
		t.Fatal("expected one rollback to fail with ErrDuplicateReversal")
	}

	result, err := walletRepository.FindByID(ctx, "wallet-concurrent-rollback-test")
	if err != nil {
		t.Fatalf("failed to find wallet: %v", err)
	}

	if result.Balance.Amount != 10000 {
		t.Fatalf("expected balance to be 10000, got %d", result.Balance.Amount)
	}

	if result.Version != 3 {
		t.Fatalf("expected wallet version 3, got %d", result.Version)
	}

	var ledgerCount int

	err = pool.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM wallet_ledger
		WHERE wallet_id = 'wallet-concurrent-rollback-test'
	`).Scan(&ledgerCount)

	if err != nil {
		t.Fatalf("failed to count ledger entries: %v", err)
	}

	if ledgerCount != 2 {
		t.Fatalf("expected exactly 2 ledger entries, got %d", ledgerCount)
	}
}

func TestPostgresWagerTransactionProcessor_RejectsDifferentReferencePlayer(t *testing.T) {
	ctx, pool, processor := setupPostgresWagerTransactionProcessorTest(t)

	now := time.Now()

	_, err := pool.Exec(ctx, `
		DELETE FROM wallet_ledger
		WHERE wallet_id = 'wallet-reference-player-test'
	`)
	if err != nil {
		t.Fatalf("failed to clean ledger: %v", err)
	}

	_, err = pool.Exec(ctx, `
		DELETE FROM wager_transactions
		WHERE id IN ('bet-reference-player-test', 'refund-reference-player-test')
	`)
	if err != nil {
		t.Fatalf("failed to clean wager transactions: %v", err)
	}

	_, err = pool.Exec(ctx, `
		DELETE FROM wallets
		WHERE id IN (
			'wallet-reference-player-test',
			'wallet-reference-player-other-test'
		)
	`)
	if err != nil {
		t.Fatalf("failed to clean wallets: %v", err)
	}

	wallet, err := domain.NewWallet(
		"wallet-reference-player-test",
		"player-reference-test",
		domain.NewMoney(10000, "BRL"),
		now,
	)
	if err != nil {
		t.Fatalf("failed to create wallet: %v", err)
	}

	walletRepository := NewPostgresWalletRepository(pool)

	if err := walletRepository.Create(ctx, wallet); err != nil {
		t.Fatalf("failed to persist wallet: %v", err)
	}

	bet, err := domain.NewWagerTransaction(
		"bet-reference-player-test",
		"provider-reference-player-test",
		"external-bet-reference-player-test",
		"idempotency-bet-reference-player-test",
		"hash-bet-reference-player-test",
		"player-reference-test",
		"wallet-reference-player-test",
		"round-reference-player-test",
		"game-reference-player-test",
		domain.TransactionBet,
		domain.NewMoney(3000, "BRL"),
		"",
		now,
	)
	if err != nil {
		t.Fatalf("failed to create bet: %v", err)
	}

	if err := processor.Execute(ctx, bet); err != nil {
		t.Fatalf("failed to process bet: %v", err)
	}

	refund, err := domain.NewWagerTransaction(
		"refund-reference-player-test",
		"provider-reference-player-test",
		"external-refund-reference-player-test",
		"idempotency-refund-reference-player-test",
		"hash-refund-reference-player-test",
		"player-other-test",
		"wallet-reference-player-test",
		"round-reference-player-test",
		"game-reference-player-test",
		domain.TransactionRefund,
		domain.NewMoney(3000, "BRL"),
		"external-bet-reference-player-test",
		now.Add(time.Second),
	)
	if err != nil {
		t.Fatalf("failed to create refund: %v", err)
	}

	err = processor.Execute(ctx, refund)

	if !errors.Is(err, ErrReferenceMismatch) {
		t.Fatalf("expected ErrReferenceMismatch, got %v", err)
	}

	result, err := walletRepository.FindByID(ctx, "wallet-reference-player-test")
	if err != nil {
		t.Fatalf("failed to find wallet: %v", err)
	}

	if result.Balance.Amount != 7000 {
		t.Fatalf("expected balance to remain 7000, got %d", result.Balance.Amount)
	}

	if result.Version != 2 {
		t.Fatalf("expected wallet version 2, got %d", result.Version)
	}
}

func TestPostgresWagerTransactionProcessor_RejectsDifferentReferenceWallet(t *testing.T) {
	ctx, pool, processor := setupPostgresWagerTransactionProcessorTest(t)

	now := time.Now()

	_, err := pool.Exec(ctx, `
		DELETE FROM wallet_ledger
		WHERE wallet_id IN (
			'wallet-reference-wallet-test',
			'wallet-reference-wallet-other-test'
		)
	`)
	if err != nil {
		t.Fatalf("failed to clean ledger: %v", err)
	}

	_, err = pool.Exec(ctx, `
		DELETE FROM wager_transactions
		WHERE id IN (
			'bet-reference-wallet-test',
			'refund-reference-wallet-test'
		)
	`)
	if err != nil {
		t.Fatalf("failed to clean wager transactions: %v", err)
	}

	_, err = pool.Exec(ctx, `
		DELETE FROM wallets
		WHERE id IN (
			'wallet-reference-wallet-test',
			'wallet-reference-wallet-other-test'
		)
	`)
	if err != nil {
		t.Fatalf("failed to clean wallets: %v", err)
	}

	walletRepository := NewPostgresWalletRepository(pool)

	wallet, err := domain.NewWallet(
		"wallet-reference-wallet-test",
		"player-reference-wallet-test",
		domain.NewMoney(10000, "BRL"),
		now,
	)
	if err != nil {
		t.Fatalf("failed to create wallet: %v", err)
	}

	if err := walletRepository.Create(ctx, wallet); err != nil {
		t.Fatalf("failed to persist wallet: %v", err)
	}

	otherWallet, err := domain.NewWallet(
		"wallet-reference-wallet-other-test",
		"player-reference-wallet-other-test",
		domain.NewMoney(10000, "BRL"),
		now,
	)

	if err != nil {
		t.Fatalf("failed to create other wallet: %v", err)
	}

	if err := walletRepository.Create(ctx, otherWallet); err != nil {
		t.Fatalf("failed to persist other wallet: %v", err)
	}

	bet, err := domain.NewWagerTransaction(
		"bet-reference-wallet-test",
		"provider-reference-wallet-test",
		"external-bet-reference-wallet-test",
		"idempotency-bet-reference-wallet-test",
		"hash-bet-reference-wallet-test",
		"player-reference-wallet-test",
		"wallet-reference-wallet-test",
		"round-reference-wallet-test",
		"game-reference-wallet-test",
		domain.TransactionBet,
		domain.NewMoney(3000, "BRL"),
		"",
		now,
	)
	if err != nil {
		t.Fatalf("failed to create bet: %v", err)
	}

	if err := processor.Execute(ctx, bet); err != nil {
		t.Fatalf("failed to process bet: %v", err)
	}

	refund, err := domain.NewWagerTransaction(
		"refund-reference-wallet-test",
		"provider-reference-wallet-test",
		"external-refund-reference-wallet-test",
		"idempotency-refund-reference-wallet-test",
		"hash-refund-reference-wallet-test",
		"player-reference-wallet-test",
		"wallet-reference-wallet-other-test",
		"round-reference-wallet-test",
		"game-reference-wallet-test",
		domain.TransactionRefund,
		domain.NewMoney(3000, "BRL"),
		"external-bet-reference-wallet-test",
		now.Add(time.Second),
	)
	if err != nil {
		t.Fatalf("failed to create refund: %v", err)
	}

	err = processor.Execute(ctx, refund)

	if !errors.Is(err, ErrReferenceMismatch) {
		t.Fatalf("expected ErrReferenceMismatch, got %v", err)
	}

	result, err := walletRepository.FindByID(
		ctx,
		"wallet-reference-wallet-test",
	)
	if err != nil {
		t.Fatalf("failed to find original wallet: %v", err)
	}

	if result.Balance.Amount != 7000 {
		t.Fatalf("expected original wallet balance to remain 7000, got %d", result.Balance.Amount)
	}

	if result.Version != 2 {
		t.Fatalf("expected original wallet version 2, got %d", result.Version)
	}

	otherResult, err := walletRepository.FindByID(
		ctx,
		"wallet-reference-wallet-other-test",
	)
	if err != nil {
		t.Fatalf("failed to find other wallet: %v", err)
	}

	if otherResult.Balance.Amount != 10000 {
		t.Fatalf("expected other wallet balance to remain 10000, got %d", otherResult.Balance.Amount)
	}

	if otherResult.Version != 1 {
		t.Fatalf("expected other wallet version 1, got %d", otherResult.Version)
	}
}

func TestPostgresWagerTransactionProcessor_RejectsDifferentReferenceRound(t *testing.T) {
	ctx, pool, processor := setupPostgresWagerTransactionProcessorTest(t)

	now := time.Now()

	_, err := pool.Exec(ctx, `
		DELETE FROM wallet_ledger
		WHERE wallet_id = 'wallet-reference-round-test'
	`)
	if err != nil {
		t.Fatalf("failed to clean ledger: %v", err)
	}

	_, err = pool.Exec(ctx, `
		DELETE FROM wager_transactions
		WHERE id IN (
			'bet-reference-round-test',
			'refund-reference-round-test'
		)
	`)
	if err != nil {
		t.Fatalf("failed to clean wager transactions: %v", err)
	}

	_, err = pool.Exec(ctx, `
		DELETE FROM wallets
		WHERE id = 'wallet-reference-round-test'
	`)
	if err != nil {
		t.Fatalf("failed to clean wallet: %v", err)
	}

	walletRepository := NewPostgresWalletRepository(pool)

	wallet, err := domain.NewWallet(
		"wallet-reference-round-test",
		"player-reference-round-test",
		domain.NewMoney(10000, "BRL"),
		now,
	)
	if err != nil {
		t.Fatalf("failed to create wallet: %v", err)
	}

	if err := walletRepository.Create(ctx, wallet); err != nil {
		t.Fatalf("failed to persist wallet: %v", err)
	}

	bet, err := domain.NewWagerTransaction(
		"bet-reference-round-test",
		"provider-reference-round-test",
		"external-bet-reference-round-test",
		"idempotency-bet-reference-round-test",
		"hash-bet-reference-round-test",
		"player-reference-round-test",
		"wallet-reference-round-test",
		"round-reference-round-test",
		"game-reference-round-test",
		domain.TransactionBet,
		domain.NewMoney(3000, "BRL"),
		"",
		now,
	)
	if err != nil {
		t.Fatalf("failed to create bet: %v", err)
	}

	if err := processor.Execute(ctx, bet); err != nil {
		t.Fatalf("failed to process bet: %v", err)
	}

	refund, err := domain.NewWagerTransaction(
		"refund-reference-round-test",
		"provider-reference-round-test",
		"external-refund-reference-round-test",
		"idempotency-refund-reference-round-test",
		"hash-refund-reference-round-test",
		"player-reference-round-test",
		"wallet-reference-round-test",
		"round-other-test",
		"game-reference-round-test",
		domain.TransactionRefund,
		domain.NewMoney(3000, "BRL"),
		"external-bet-reference-round-test",
		now.Add(time.Second),
	)
	if err != nil {
		t.Fatalf("failed to create refund: %v", err)
	}

	err = processor.Execute(ctx, refund)

	if !errors.Is(err, ErrReferenceMismatch) {
		t.Fatalf("expected ErrReferenceMismatch, got %v", err)
	}

	result, err := walletRepository.FindByID(
		ctx,
		"wallet-reference-round-test",
	)
	if err != nil {
		t.Fatalf("failed to find wallet: %v", err)
	}

	if result.Balance.Amount != 7000 {
		t.Fatalf("expected balance 7000, got %d", result.Balance.Amount)
	}

	if result.Version != 2 {
		t.Fatalf("expected version 2, got %d", result.Version)
	}
}

func TestPostgresWagerTransactionProcessor_RejectsDifferentReferenceGame(t *testing.T) {
	ctx, pool, processor := setupPostgresWagerTransactionProcessorTest(t)

	now := time.Now()

	_, err := pool.Exec(ctx, `
		DELETE FROM wallet_ledger
		WHERE wallet_id = 'wallet-reference-game-test'
	`)
	if err != nil {
		t.Fatalf("failed to clean ledger: %v", err)
	}

	_, err = pool.Exec(ctx, `
		DELETE FROM wager_transactions
		WHERE id IN (
			'bet-reference-game-test',
			'refund-reference-game-test'
		)
	`)
	if err != nil {
		t.Fatalf("failed to clean wager transactions: %v", err)
	}

	_, err = pool.Exec(ctx, `
		DELETE FROM wallets
		WHERE id = 'wallet-reference-game-test'
	`)
	if err != nil {
		t.Fatalf("failed to clean wallet: %v", err)
	}

	walletRepository := NewPostgresWalletRepository(pool)

	wallet, err := domain.NewWallet(
		"wallet-reference-game-test",
		"player-reference-game-test",
		domain.NewMoney(10000, "BRL"),
		now,
	)
	if err != nil {
		t.Fatalf("failed to create wallet: %v", err)
	}

	if err := walletRepository.Create(ctx, wallet); err != nil {
		t.Fatalf("failed to persist wallet: %v", err)
	}

	bet, err := domain.NewWagerTransaction(
		"bet-reference-game-test",
		"provider-reference-game-test",
		"external-bet-reference-game-test",
		"idempotency-bet-reference-game-test",
		"hash-bet-reference-game-test",
		"player-reference-game-test",
		"wallet-reference-game-test",
		"round-reference-game-test",
		"game-reference-game-test",
		domain.TransactionBet,
		domain.NewMoney(3000, "BRL"),
		"",
		now,
	)
	if err != nil {
		t.Fatalf("failed to create bet: %v", err)
	}

	if err := processor.Execute(ctx, bet); err != nil {
		t.Fatalf("failed to process bet: %v", err)
	}

	refund, err := domain.NewWagerTransaction(
		"refund-reference-game-test",
		"provider-reference-game-test",
		"external-refund-reference-game-test",
		"idempotency-refund-reference-game-test",
		"hash-refund-reference-game-test",
		"player-reference-game-test",
		"wallet-reference-game-test",
		"round-reference-game-test",
		"game-other-test",
		domain.TransactionRefund,
		domain.NewMoney(3000, "BRL"),
		"external-bet-reference-game-test",
		now.Add(time.Second),
	)
	if err != nil {
		t.Fatalf("failed to create refund: %v", err)
	}

	err = processor.Execute(ctx, refund)

	if !errors.Is(err, ErrReferenceMismatch) {
		t.Fatalf("expected ErrReferenceMismatch, got %v", err)
	}

	result, err := walletRepository.FindByID(
		ctx,
		"wallet-reference-game-test",
	)
	if err != nil {
		t.Fatalf("failed to find wallet: %v", err)
	}

	if result.Balance.Amount != 7000 {
		t.Fatalf("expected balance 7000, got %d", result.Balance.Amount)
	}

	if result.Version != 2 {
		t.Fatalf("expected version 2, got %d", result.Version)
	}
}

func TestPostgresWagerTransactionProcessor_LossDoesNotChangeBalance(t *testing.T) {
	ctx, pool, processor := setupPostgresWagerTransactionProcessorTest(t)

	now := time.Now()

	_, err := pool.Exec(ctx, `
		DELETE FROM wallet_ledger
		WHERE wallet_id = 'wallet-loss-test'
	`)
	if err != nil {
		t.Fatalf("failed to clean ledger: %v", err)
	}

	_, err = pool.Exec(ctx, `
		DELETE FROM wager_transactions
		WHERE id = 'transaction-loss-test'
	`)
	if err != nil {
		t.Fatalf("failed to clean wager transaction: %v", err)
	}

	_, err = pool.Exec(ctx, `
		DELETE FROM wallets
		WHERE id = 'wallet-loss-test'
	`)
	if err != nil {
		t.Fatalf("failed to clean wallet: %v", err)
	}

	walletRepository := NewPostgresWalletRepository(pool)
	transactionRepository := NewPostgresWagerTransactionRepository(pool)

	wallet, err := domain.NewWallet(
		"wallet-loss-test",
		"player-loss-test",
		domain.NewMoney(10000, "BRL"),
		now,
	)
	if err != nil {
		t.Fatalf("failed to create wallet: %v", err)
	}

	if err := walletRepository.Create(ctx, wallet); err != nil {
		t.Fatalf("failed to persist wallet: %v", err)
	}

	transaction, err := domain.NewWagerTransaction(
		"transaction-loss-test",
		"provider-loss-test",
		"external-loss-test",
		"idempotency-loss-test",
		"hash-loss-test",
		"player-loss-test",
		"wallet-loss-test",
		"round-loss-test",
		"game-loss-test",
		domain.TransactionLoss,
		domain.NewMoney(0, "BRL"),
		"",
		now,
	)
	if err != nil {
		t.Fatalf("failed to create loss transaction: %v", err)
	}

	if err := processor.Execute(ctx, transaction); err != nil {
		t.Fatalf("failed to process loss transaction: %v", err)
	}

	result, err := walletRepository.FindByID(ctx, "wallet-loss-test")
	if err != nil {
		t.Fatalf("failed to find wallet: %v", err)
	}

	if result.Balance.Amount != 10000 {
		t.Fatalf("expected balance 10000, got %d", result.Balance.Amount)
	}

	if result.Version != 1 {
		t.Fatalf("expected version 1, got %d", result.Version)
	}

	processed, err := transactionRepository.FindByID(
		ctx,
		"transaction-loss-test",
	)
	if err != nil {
		t.Fatalf("failed to find loss transaction: %v", err)
	}

	if processed.State != domain.TransactionProcessed {
		t.Fatalf("expected state PROCESSED, got %s", processed.State)
	}

	var ledgerCount int

	err = pool.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM wallet_ledger
		WHERE wallet_id = 'wallet-loss-test'
	`).Scan(&ledgerCount)

	if err != nil {
		t.Fatalf("failed to count ledger entries: %v", err)
	}

	if ledgerCount != 0 {
		t.Fatalf("expected no ledger entries for LOSS, got %d", ledgerCount)
	}
}

func TestPostgresWagerTransactionProcessor_IdempotentBet(t *testing.T) {
	ctx, pool, processor := setupPostgresWagerTransactionProcessorTest(t)

	now := time.Now()

	_, err := pool.Exec(ctx, `
		DELETE FROM wallet_ledger
		WHERE wallet_id = 'wallet-idempotent-bet-test'
	`)
	if err != nil {
		t.Fatalf("failed to clean ledger: %v", err)
	}

	_, err = pool.Exec(ctx, `
		DELETE FROM wager_transactions
		WHERE id = 'transaction-idempotent-bet-test'
	`)
	if err != nil {
		t.Fatalf("failed to clean wager transaction: %v", err)
	}

	_, err = pool.Exec(ctx, `
		DELETE FROM wallets
		WHERE id = 'wallet-idempotent-bet-test'
	`)
	if err != nil {
		t.Fatalf("failed to clean wallet: %v", err)
	}

	walletRepository := NewPostgresWalletRepository(pool)
	transactionRepository := NewPostgresWagerTransactionRepository(pool)

	wallet, err := domain.NewWallet(
		"wallet-idempotent-bet-test",
		"player-idempotent-bet-test",
		domain.NewMoney(10000, "BRL"),
		now,
	)
	if err != nil {
		t.Fatalf("failed to create wallet: %v", err)
	}

	if err := walletRepository.Create(ctx, wallet); err != nil {
		t.Fatalf("failed to persist wallet: %v", err)
	}

	transaction, err := domain.NewWagerTransaction(
		"transaction-idempotent-bet-test",
		"provider-idempotent-bet-test",
		"external-idempotent-bet-test",
		"idempotency-idempotent-bet-test",
		"hash-idempotent-bet-test",
		"player-idempotent-bet-test",
		"wallet-idempotent-bet-test",
		"round-idempotent-bet-test",
		"game-idempotent-bet-test",
		domain.TransactionBet,
		domain.NewMoney(3000, "BRL"),
		"",
		now,
	)
	if err != nil {
		t.Fatalf("failed to create transaction: %v", err)
	}

	// Primeira execução.
	if err := processor.Execute(ctx, transaction); err != nil {
		t.Fatalf("first execution failed: %v", err)
	}

	// Segunda execução da mesma transação.
	if err := processor.Execute(ctx, transaction); err != nil {
		t.Fatalf("second execution failed: %v", err)
	}

	result, err := walletRepository.FindByID(
		ctx,
		"wallet-idempotent-bet-test",
	)
	if err != nil {
		t.Fatalf("failed to find wallet: %v", err)
	}

	if result.Balance.Amount != 7000 {
		t.Fatalf("expected balance 7000, got %d", result.Balance.Amount)
	}

	if result.Version != 2 {
		t.Fatalf("expected version 2, got %d", result.Version)
	}

	processed, err := transactionRepository.FindByID(
		ctx,
		"transaction-idempotent-bet-test",
	)
	if err != nil {
		t.Fatalf("failed to find transaction: %v", err)
	}

	if processed.State != domain.TransactionProcessed {
		t.Fatalf("expected state PROCESSED, got %s", processed.State)
	}

	var ledgerCount int

	err = pool.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM wallet_ledger
		WHERE wallet_id = 'wallet-idempotent-bet-test'
		AND transaction_id = 'transaction-idempotent-bet-test'
	`).Scan(&ledgerCount)

	if err != nil {
		t.Fatalf("failed to count ledger entries: %v", err)
	}

	if ledgerCount != 1 {
		t.Fatalf("expected exactly 1 ledger entry, got %d", ledgerCount)
	}
}

func TestPostgresWagerTransactionProcessor_IdempotencyConflict(t *testing.T) {
	ctx, pool, processor := setupPostgresWagerTransactionProcessorTest(t)

	now := time.Now()

	_, err := pool.Exec(ctx, `
		DELETE FROM wallet_ledger
		WHERE wallet_id = 'wallet-idempotency-conflict-test'
	`)
	if err != nil {
		t.Fatalf("failed to clean ledger: %v", err)
	}

	_, err = pool.Exec(ctx, `
		DELETE FROM wager_transactions
		WHERE id IN (
			'transaction-idempotency-conflict-1',
			'transaction-idempotency-conflict-2'
		)
	`)
	if err != nil {
		t.Fatalf("failed to clean wager transactions: %v", err)
	}

	_, err = pool.Exec(ctx, `
		DELETE FROM wallets
		WHERE id = 'wallet-idempotency-conflict-test'
	`)
	if err != nil {
		t.Fatalf("failed to clean wallet: %v", err)
	}

	walletRepository := NewPostgresWalletRepository(pool)

	wallet, err := domain.NewWallet(
		"wallet-idempotency-conflict-test",
		"player-idempotency-conflict-test",
		domain.NewMoney(10000, "BRL"),
		now,
	)
	if err != nil {
		t.Fatalf("failed to create wallet: %v", err)
	}

	if err := walletRepository.Create(ctx, wallet); err != nil {
		t.Fatalf("failed to persist wallet: %v", err)
	}

	first, err := domain.NewWagerTransaction(
		"transaction-idempotency-conflict-1",
		"provider-idempotency-conflict-test",
		"external-idempotency-conflict-1",
		"idempotency-conflict-key",
		"hash-payload-1",
		"player-idempotency-conflict-test",
		"wallet-idempotency-conflict-test",
		"round-idempotency-conflict-test",
		"game-idempotency-conflict-test",
		domain.TransactionBet,
		domain.NewMoney(3000, "BRL"),
		"",
		now,
	)
	if err != nil {
		t.Fatalf("failed to create first transaction: %v", err)
	}

	if err := processor.Execute(ctx, first); err != nil {
		t.Fatalf("failed to process first transaction: %v", err)
	}

	second, err := domain.NewWagerTransaction(
		"transaction-idempotency-conflict-2",
		"provider-idempotency-conflict-test",
		"external-idempotency-conflict-2",
		"idempotency-conflict-key",
		"hash-payload-2",
		"player-idempotency-conflict-test",
		"wallet-idempotency-conflict-test",
		"round-idempotency-conflict-test",
		"game-idempotency-conflict-test",
		domain.TransactionBet,
		domain.NewMoney(2000, "BRL"),
		"",
		now.Add(time.Second),
	)
	if err != nil {
		t.Fatalf("failed to create second transaction: %v", err)
	}

	err = processor.Execute(ctx, second)

	if !errors.Is(err, ErrIdempotencyConflict) {
		t.Fatalf("expected ErrIdempotencyConflict, got %v", err)
	}

	result, err := walletRepository.FindByID(
		ctx,
		"wallet-idempotency-conflict-test",
	)
	if err != nil {
		t.Fatalf("failed to find wallet: %v", err)
	}

	if result.Balance.Amount != 7000 {
		t.Fatalf("expected balance 7000, got %d", result.Balance.Amount)
	}

	if result.Version != 2 {
		t.Fatalf("expected version 2, got %d", result.Version)
	}
}

func TestPostgresWagerTransactionProcessor_ProviderTransactionConflict(t *testing.T) {
	ctx, pool, processor := setupPostgresWagerTransactionProcessorTest(t)

	now := time.Now()

	_, err := pool.Exec(ctx, `
		DELETE FROM wallet_ledger
		WHERE wallet_id = 'wallet-provider-conflict-test'
	`)
	if err != nil {
		t.Fatalf("failed to clean ledger: %v", err)
	}

	_, err = pool.Exec(ctx, `
		DELETE FROM wager_transactions
		WHERE id IN (
			'transaction-provider-conflict-1',
			'transaction-provider-conflict-2'
		)
	`)
	if err != nil {
		t.Fatalf("failed to clean wager transactions: %v", err)
	}

	_, err = pool.Exec(ctx, `
		DELETE FROM wallets
		WHERE id = 'wallet-provider-conflict-test'
	`)
	if err != nil {
		t.Fatalf("failed to clean wallet: %v", err)
	}

	walletRepository := NewPostgresWalletRepository(pool)

	wallet, err := domain.NewWallet(
		"wallet-provider-conflict-test",
		"player-provider-conflict-test",
		domain.NewMoney(10000, "BRL"),
		now,
	)
	if err != nil {
		t.Fatalf("failed to create wallet: %v", err)
	}

	if err := walletRepository.Create(ctx, wallet); err != nil {
		t.Fatalf("failed to persist wallet: %v", err)
	}

	first, err := domain.NewWagerTransaction(
		"transaction-provider-conflict-1",
		"provider-provider-conflict-test",
		"external-provider-conflict-test",
		"idempotency-provider-conflict-1",
		"hash-provider-conflict-1",
		"player-provider-conflict-test",
		"wallet-provider-conflict-test",
		"round-provider-conflict-test",
		"game-provider-conflict-test",
		domain.TransactionBet,
		domain.NewMoney(3000, "BRL"),
		"",
		now,
	)
	if err != nil {
		t.Fatalf("failed to create first transaction: %v", err)
	}

	if err := processor.Execute(ctx, first); err != nil {
		t.Fatalf("failed to process first transaction: %v", err)
	}

	second, err := domain.NewWagerTransaction(
		"transaction-provider-conflict-2",
		"provider-provider-conflict-test",
		"external-provider-conflict-test",
		"idempotency-provider-conflict-2",
		"hash-provider-conflict-2",
		"player-provider-conflict-test",
		"wallet-provider-conflict-test",
		"round-provider-conflict-test",
		"game-provider-conflict-test",
		domain.TransactionBet,
		domain.NewMoney(3000, "BRL"),
		"",
		now.Add(time.Second),
	)
	if err != nil {
		t.Fatalf("failed to create second transaction: %v", err)
	}

	err = processor.Execute(ctx, second)

	if !errors.Is(err, ErrProviderTransactionConflict) {
		t.Fatalf("expected ErrProviderTransactionConflict, got %v", err)
	}

	result, err := walletRepository.FindByID(
		ctx,
		"wallet-provider-conflict-test",
	)
	if err != nil {
		t.Fatalf("failed to find wallet: %v", err)
	}

	if result.Balance.Amount != 7000 {
		t.Fatalf("expected balance 7000, got %d", result.Balance.Amount)
	}

	if result.Version != 2 {
		t.Fatalf("expected version 2, got %d", result.Version)
	}
}

func TestPostgresWagerTransactionProcessor_ConcurrentSameIdempotencyKey(t *testing.T) {
	ctx, pool, processor := setupPostgresWagerTransactionProcessorTest(t)

	now := time.Now()

	_, err := pool.Exec(ctx, `
		DELETE FROM wallet_ledger
		WHERE wallet_id = 'wallet-concurrent-idempotency-test'
	`)
	if err != nil {
		t.Fatalf("failed to clean ledger: %v", err)
	}

	_, err = pool.Exec(ctx, `
		DELETE FROM wager_transactions
		WHERE provider_id = 'provider-concurrent-idempotency-test'
	`)
	if err != nil {
		t.Fatalf("failed to clean wager transactions: %v", err)
	}

	_, err = pool.Exec(ctx, `
		DELETE FROM wallets
		WHERE id = 'wallet-concurrent-idempotency-test'
	`)
	if err != nil {
		t.Fatalf("failed to clean wallet: %v", err)
	}

	walletRepository := NewPostgresWalletRepository(pool)

	wallet, err := domain.NewWallet(
		"wallet-concurrent-idempotency-test",
		"player-concurrent-idempotency-test",
		domain.NewMoney(10000, "BRL"),
		now,
	)
	if err != nil {
		t.Fatalf("failed to create wallet: %v", err)
	}

	if err := walletRepository.Create(ctx, wallet); err != nil {
		t.Fatalf("failed to persist wallet: %v", err)
	}

	createTransaction := func() domain.WagerTransaction {
		transaction, err := domain.NewWagerTransaction(
			"transaction-concurrent-idempotency-test",
			"provider-concurrent-idempotency-test",
			"external-concurrent-idempotency-test",
			"idempotency-concurrent-idempotency-test",
			"hash-concurrent-idempotency-test",
			"player-concurrent-idempotency-test",
			"wallet-concurrent-idempotency-test",
			"round-concurrent-idempotency-test",
			"game-concurrent-idempotency-test",
			domain.TransactionBet,
			domain.NewMoney(3000, "BRL"),
			"",
			now,
		)
		if err != nil {
			t.Fatalf("failed to create transaction: %v", err)
		}

		return transaction
	}

	var wg sync.WaitGroup
	errs := make(chan error, 2)

	wg.Add(2)

	go func() {
		defer wg.Done()
		errs <- processor.Execute(ctx, createTransaction())
	}()

	go func() {
		defer wg.Done()
		errs <- processor.Execute(ctx, createTransaction())
	}()

	wg.Wait()
	close(errs)

	for err := range errs {
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}

	result, err := walletRepository.FindByID(
		ctx,
		"wallet-concurrent-idempotency-test",
	)
	if err != nil {
		t.Fatalf("failed to find wallet: %v", err)
	}

	if result.Balance.Amount != 7000 {
		t.Fatalf("expected balance 7000, got %d", result.Balance.Amount)
	}

	if result.Version != 2 {
		t.Fatalf("expected version 2, got %d", result.Version)
	}

	var transactionCount int

	err = pool.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM wager_transactions
		WHERE idempotency_key = 'idempotency-concurrent-idempotency-test'
	`).Scan(&transactionCount)

	if err != nil {
		t.Fatalf("failed to count transactions: %v", err)
	}

	if transactionCount != 1 {
		t.Fatalf("expected exactly 1 transaction, got %d", transactionCount)
	}

	var ledgerCount int

	err = pool.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM wallet_ledger
		WHERE wallet_id = 'wallet-concurrent-idempotency-test'
	`).Scan(&ledgerCount)

	if err != nil {
		t.Fatalf("failed to count ledger entries: %v", err)
	}

	if ledgerCount != 1 {
		t.Fatalf("expected exactly 1 ledger entry, got %d", ledgerCount)
	}
}

func TestPostgresWagerTransactionProcessor_ResumePendingReference(t *testing.T) {
	ctx, pool, processor := setupPostgresWagerTransactionProcessorTest(t)

	now := time.Now()

	_, err := pool.Exec(ctx, `
		DELETE FROM wallet_ledger
		WHERE wallet_id = 'wallet-resume-reference-test'
	`)
	if err != nil {
		t.Fatalf("failed to clean ledger: %v", err)
	}

	_, err = pool.Exec(ctx, `
		DELETE FROM wager_transactions
		WHERE provider_id = 'provider-resume-reference-test'
	`)
	if err != nil {
		t.Fatalf("failed to clean wager transactions: %v", err)
	}

	_, err = pool.Exec(ctx, `
		DELETE FROM wallets
		WHERE id = 'wallet-resume-reference-test'
	`)
	if err != nil {
		t.Fatalf("failed to clean wallet: %v", err)
	}

	walletRepository := NewPostgresWalletRepository(pool)
	wagerTransactionRepository := NewPostgresWagerTransactionRepository(pool)

	wallet, err := domain.NewWallet(
		"wallet-resume-reference-test",
		"player-resume-reference-test",
		domain.NewMoney(10000, "BRL"),
		now,
	)
	if err != nil {
		t.Fatalf("failed to create wallet: %v", err)
	}

	if err := walletRepository.Create(ctx, wallet); err != nil {
		t.Fatalf("failed to persist wallet: %v", err)
	}

	reference, err := domain.NewWagerTransaction(
		"reference-resume-reference-test",
		"provider-resume-reference-test",
		"external-reference-resume-reference-test",
		"idempotency-reference-resume-reference-test",
		"hash-reference-resume-reference-test",
		"player-resume-reference-test",
		"wallet-resume-reference-test",
		"round-resume-reference-test",
		"game-resume-reference-test",
		domain.TransactionBet,
		domain.NewMoney(3000, "BRL"),
		"",
		now,
	)
	if err != nil {
		t.Fatalf("failed to create reference transaction: %v", err)
	}

	refund, err := domain.NewWagerTransaction(
		"refund-resume-reference-test",
		"provider-resume-reference-test",
		"external-refund-resume-reference-test",
		"idempotency-refund-resume-reference-test",
		"hash-refund-resume-reference-test",
		"player-resume-reference-test",
		"wallet-resume-reference-test",
		"round-resume-reference-test",
		"game-resume-reference-test",
		domain.TransactionRefund,
		domain.NewMoney(3000, "BRL"),
		"external-reference-resume-reference-test",
		now,
	)
	if err != nil {
		t.Fatalf("failed to create refund transaction: %v", err)
	}

	if err := processor.Execute(ctx, refund); err != nil {
		t.Fatalf("failed to create pending refund: %v", err)
	}

	pendingRefund, err := wagerTransactionRepository.FindByID(
		ctx,
		"refund-resume-reference-test",
	)
	if err != nil {
		t.Fatalf("failed to find pending refund: %v", err)
	}

	if pendingRefund.State != domain.TransactionPendingReference {
		t.Fatalf(
			"expected refund state PENDING_REFERENCE, got %s",
			pendingRefund.State,
		)
	}

	if err := processor.Execute(ctx, reference); err != nil {
		t.Fatalf("failed to process reference transaction: %v", err)
	}

	if err := processor.ResumePendingReference(
		ctx,
		"refund-resume-reference-test",
	); err != nil {
		t.Fatalf("failed to resume pending reference: %v", err)
	}

	resumedRefund, err := wagerTransactionRepository.FindByID(
		ctx,
		"refund-resume-reference-test",
	)
	if err != nil {
		t.Fatalf("failed to find resumed refund: %v", err)
	}

	if resumedRefund.State != domain.TransactionProcessed {
		t.Fatalf(
			"expected refund state PROCESSED, got %s",
			resumedRefund.State,
		)
	}

	result, err := walletRepository.FindByID(
		ctx,
		"wallet-resume-reference-test",
	)
	if err != nil {
		t.Fatalf("failed to find wallet: %v", err)
	}

	if result.Balance.Amount != 10000 {
		t.Fatalf(
			"expected balance 10000, got %d",
			result.Balance.Amount,
		)
	}

	if result.Version != 3 {
		t.Fatalf(
			"expected wallet version 3, got %d",
			result.Version,
		)
	}

	var ledgerCount int

	err = pool.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM wallet_ledger
		WHERE wallet_id = 'wallet-resume-reference-test'
	`).Scan(&ledgerCount)

	if err != nil {
		t.Fatalf("failed to count ledger entries: %v", err)
	}

	if ledgerCount != 2 {
		t.Fatalf(
			"expected exactly 2 ledger entries, got %d",
			ledgerCount,
		)
	}
}
