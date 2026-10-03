package application

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/yasto/backend-challenge/internal/domain"
)

func TestProcessWagerTransactionBet(t *testing.T) {
	repository := NewInMemoryWalletRepository()

	wallet, err := domain.NewWallet(
		"wallet-123",
		"player-123",
		domain.NewMoney(10000, "BRL"),
		time.Now(),
	)
	if err != nil {
		t.Fatalf("unexpected error creating wallet: %v", err)
	}

	if err := repository.Save(context.Background(), wallet); err != nil {
		t.Fatalf("unexpected error saving wallet: %v", err)
	}

	transaction, err := domain.NewWagerTransaction(
		"transaction-123",
		"provider-123",
		"external-123",
		"idempotency-123",
		"hash-123",
		"player-123",
		"wallet-123",
		"round-123",
		"game-123",
		domain.TransactionBet,
		domain.NewMoney(3000, "BRL"),
		"",
		time.Now(),
	)
	if err != nil {
		t.Fatalf("unexpected error creating transaction: %v", err)
	}

	ledgerRepository := NewInMemoryLedgerRepository()

	useCase := NewProcessWagerTransaction(
		repository,
		ledgerRepository,
	)

	err = useCase.Execute(context.Background(), transaction)
	if err != nil {
		t.Fatalf("unexpected error processing transaction: %v", err)
	}

	updatedWallet, err := repository.FindByID(
		context.Background(),
		"wallet-123",
	)
	if err != nil {
		t.Fatalf("unexpected error finding wallet: %v", err)
	}

	if updatedWallet.Balance.Amount != 7000 {
		t.Fatalf(
			"expected balance 7000, got %d",
			updatedWallet.Balance.Amount,
		)
	}

	if updatedWallet.Version != 2 {
		t.Fatalf(
			"expected wallet version 2, got %d",
			updatedWallet.Version,
		)
	}

	entries := ledgerRepository.Entries()

	if len(entries) != 1 {
		t.Fatalf("expected 1 ledger entry, got %d", len(entries))
	}

	entry := entries[0]

	if entry.Direction != domain.LedgerDebit {
		t.Fatalf("expected DEBIT, got %s", entry.Direction)
	}

	if entry.Amount.Amount != 3000 {
		t.Fatalf(
			"expected ledger amount 3000, got %d",
			entry.Amount.Amount,
		)
	}

	if entry.BalanceBefore.Amount != 10000 {
		t.Fatalf(
			"expected balance before 10000, got %d",
			entry.BalanceBefore.Amount,
		)
	}

	if entry.BalanceAfter.Amount != 7000 {
		t.Fatalf(
			"expected balance after 7000, got %d",
			entry.BalanceAfter.Amount,
		)
	}
}

func TestProcessWagerTransactionWin(t *testing.T) {
	repository := NewInMemoryWalletRepository()
	ledgerRepository := NewInMemoryLedgerRepository()

	wallet, err := domain.NewWallet(
		"wallet-123",
		"player-123",
		domain.NewMoney(10000, "BRL"),
		time.Now(),
	)
	if err != nil {
		t.Fatalf("unexpected error creating wallet: %v", err)
	}

	if err := repository.Save(context.Background(), wallet); err != nil {
		t.Fatalf("unexpected error saving wallet: %v", err)
	}

	transaction, err := domain.NewWagerTransaction(
		"transaction-win-123",
		"provider-123",
		"external-win-123",
		"idempotency-win-123",
		"hash-win-123",
		"player-123",
		"wallet-123",
		"round-123",
		"game-123",
		domain.TransactionWin,
		domain.NewMoney(3000, "BRL"),
		"",
		time.Now(),
	)
	if err != nil {
		t.Fatalf("unexpected error creating transaction: %v", err)
	}

	useCase := NewProcessWagerTransaction(
		repository,
		ledgerRepository,
	)

	err = useCase.Execute(context.Background(), transaction)
	if err != nil {
		t.Fatalf("unexpected error processing transaction: %v", err)
	}

	updatedWallet, err := repository.FindByID(
		context.Background(),
		"wallet-123",
	)
	if err != nil {
		t.Fatalf("unexpected error finding wallet: %v", err)
	}

	if updatedWallet.Balance.Amount != 13000 {
		t.Fatalf(
			"expected balance 13000, got %d",
			updatedWallet.Balance.Amount,
		)
	}

	if updatedWallet.Version != 2 {
		t.Fatalf(
			"expected wallet version 2, got %d",
			updatedWallet.Version,
		)
	}

	entries := ledgerRepository.Entries()

	if len(entries) != 1 {
		t.Fatalf("expected 1 ledger entry, got %d", len(entries))
	}

	entry := entries[0]

	if entry.Direction != domain.LedgerCredit {
		t.Fatalf("expected CREDIT, got %s", entry.Direction)
	}

	if entry.Amount.Amount != 3000 {
		t.Fatalf(
			"expected ledger amount 3000, got %d",
			entry.Amount.Amount,
		)
	}

	if entry.BalanceBefore.Amount != 10000 {
		t.Fatalf(
			"expected balance before 10000, got %d",
			entry.BalanceBefore.Amount,
		)
	}

	if entry.BalanceAfter.Amount != 13000 {
		t.Fatalf(
			"expected balance after 13000, got %d",
			entry.BalanceAfter.Amount,
		)
	}
}

func TestProcessWagerTransactionLoss(t *testing.T) {
	repository := NewInMemoryWalletRepository()
	ledgerRepository := NewInMemoryLedgerRepository()

	wallet, err := domain.NewWallet(
		"wallet-loss-123",
		"player-123",
		domain.NewMoney(10000, "BRL"),
		time.Now(),
	)
	if err != nil {
		t.Fatalf("unexpected error creating wallet: %v", err)
	}

	if err := repository.Save(context.Background(), wallet); err != nil {
		t.Fatalf("unexpected error saving wallet: %v", err)
	}

	transaction, err := domain.NewWagerTransaction(
		"transaction-loss-123",
		"provider-123",
		"external-loss-123",
		"idempotency-loss-123",
		"hash-loss-123",
		"player-123",
		"wallet-loss-123",
		"round-123",
		"game-123",
		domain.TransactionLoss,
		domain.NewMoney(0, "BRL"),
		"",
		time.Now(),
	)
	if err != nil {
		t.Fatalf("unexpected error creating transaction: %v", err)
	}

	useCase := NewProcessWagerTransaction(
		repository,
		ledgerRepository,
	)

	err = useCase.Execute(context.Background(), transaction)
	if err != nil {
		t.Fatalf("unexpected error processing transaction: %v", err)
	}

	updatedWallet, err := repository.FindByID(
		context.Background(),
		"wallet-loss-123",
	)
	if err != nil {
		t.Fatalf("unexpected error finding wallet: %v", err)
	}

	if updatedWallet.Balance.Amount != 10000 {
		t.Fatalf(
			"expected balance 10000, got %d",
			updatedWallet.Balance.Amount,
		)
	}

	if updatedWallet.Version != 1 {
		t.Fatalf(
			"expected wallet version 1, got %d",
			updatedWallet.Version,
		)
	}

	entries := ledgerRepository.Entries()

	if len(entries) != 0 {
		t.Fatalf(
			"expected 0 ledger entries, got %d",
			len(entries),
		)
	}
}

func TestProcessWagerTransactionRefund(t *testing.T) {
	repository := NewInMemoryWalletRepository()
	ledgerRepository := NewInMemoryLedgerRepository()

	wallet, err := domain.NewWallet(
		"wallet-refund-123",
		"player-123",
		domain.NewMoney(10000, "BRL"),
		time.Now(),
	)
	if err != nil {
		t.Fatalf("unexpected error creating wallet: %v", err)
	}

	if err := repository.Save(context.Background(), wallet); err != nil {
		t.Fatalf("unexpected error saving wallet: %v", err)
	}

	// BET original de 30 BRL.
	betTransaction, err := domain.NewWagerTransaction(
		"transaction-bet-refund-123",
		"provider-123",
		"external-bet-refund-123",
		"idempotency-bet-refund-123",
		"hash-bet-refund-123",
		"player-123",
		"wallet-refund-123",
		"round-123",
		"game-123",
		domain.TransactionBet,
		domain.NewMoney(3000, "BRL"),
		"",
		time.Now(),
	)
	if err != nil {
		t.Fatalf("unexpected error creating BET: %v", err)
	}

	useCase := NewProcessWagerTransaction(
		repository,
		ledgerRepository,
	)

	if err := useCase.Execute(
		context.Background(),
		betTransaction,
	); err != nil {
		t.Fatalf("unexpected error processing BET: %v", err)
	}

	// REFUND do BET original.
	refundTransaction, err := domain.NewWagerTransaction(
		"transaction-refund-123",
		"provider-123",
		"external-refund-123",
		"idempotency-refund-123",
		"hash-refund-123",
		"player-123",
		"wallet-refund-123",
		"round-123",
		"game-123",
		domain.TransactionRefund,
		domain.NewMoney(3000, "BRL"),
		"external-bet-refund-123",
		time.Now(),
	)
	if err != nil {
		t.Fatalf("unexpected error creating REFUND: %v", err)
	}

	if err := useCase.Execute(
		context.Background(),
		refundTransaction,
	); err != nil {
		t.Fatalf("unexpected error processing REFUND: %v", err)
	}

	updatedWallet, err := repository.FindByID(
		context.Background(),
		"wallet-refund-123",
	)
	if err != nil {
		t.Fatalf("unexpected error finding wallet: %v", err)
	}

	// 100 - 30 + 30 = 100.
	if updatedWallet.Balance.Amount != 10000 {
		t.Fatalf(
			"expected balance 10000, got %d",
			updatedWallet.Balance.Amount,
		)
	}

	// BET alterou a versão e REFUND também.
	if updatedWallet.Version != 3 {
		t.Fatalf(
			"expected wallet version 3, got %d",
			updatedWallet.Version,
		)
	}

	entries := ledgerRepository.Entries()

	if len(entries) != 2 {
		t.Fatalf(
			"expected 2 ledger entries, got %d",
			len(entries),
		)
	}

	refundEntry := entries[1]

	if refundEntry.Direction != domain.LedgerCredit {
		t.Fatalf(
			"expected REFUND ledger direction CREDIT, got %s",
			refundEntry.Direction,
		)
	}

	if refundEntry.Amount.Amount != 3000 {
		t.Fatalf(
			"expected refund amount 3000, got %d",
			refundEntry.Amount.Amount,
		)
	}
}

func TestProcessWagerTransactionRollbackBet(t *testing.T) {
	repository := NewInMemoryWalletRepository()
	ledgerRepository := NewInMemoryLedgerRepository()

	wallet, err := domain.NewWallet(
		"wallet-rollback-123",
		"player-123",
		domain.NewMoney(10000, "BRL"),
		time.Now(),
	)
	if err != nil {
		t.Fatalf("unexpected error creating wallet: %v", err)
	}

	if err := repository.Save(context.Background(), wallet); err != nil {
		t.Fatalf("unexpected error saving wallet: %v", err)
	}

	useCase := NewProcessWagerTransaction(
		repository,
		ledgerRepository,
	)

	betTransaction, err := domain.NewWagerTransaction(
		"transaction-bet-rollback-123",
		"provider-123",
		"external-bet-rollback-123",
		"idempotency-bet-rollback-123",
		"hash-bet-rollback-123",
		"player-123",
		"wallet-rollback-123",
		"round-123",
		"game-123",
		domain.TransactionBet,
		domain.NewMoney(3000, "BRL"),
		"",
		time.Now(),
	)
	if err != nil {
		t.Fatalf("unexpected error creating BET: %v", err)
	}

	if err := useCase.Execute(
		context.Background(),
		betTransaction,
	); err != nil {
		t.Fatalf("unexpected error processing BET: %v", err)
	}

	rollbackTransaction, err := domain.NewWagerTransaction(
		"transaction-rollback-123",
		"provider-123",
		"external-rollback-123",
		"idempotency-rollback-123",
		"hash-rollback-123",
		"player-123",
		"wallet-rollback-123",
		"round-123",
		"game-123",
		domain.TransactionRollback,
		domain.NewMoney(3000, "BRL"),
		"external-bet-rollback-123",
		time.Now(),
	)
	if err != nil {
		t.Fatalf("unexpected error creating ROLLBACK: %v", err)
	}

	if err := useCase.Execute(
		context.Background(),
		rollbackTransaction,
	); err != nil {
		t.Fatalf("unexpected error processing ROLLBACK: %v", err)
	}

	updatedWallet, err := repository.FindByID(
		context.Background(),
		"wallet-rollback-123",
	)
	if err != nil {
		t.Fatalf("unexpected error finding wallet: %v", err)
	}

	// 100 - 30 + 30 = 100.
	if updatedWallet.Balance.Amount != 10000 {
		t.Fatalf(
			"expected balance 10000, got %d",
			updatedWallet.Balance.Amount,
		)
	}

	if updatedWallet.Version != 3 {
		t.Fatalf(
			"expected wallet version 3, got %d",
			updatedWallet.Version,
		)
	}

	entries := ledgerRepository.Entries()

	if len(entries) != 2 {
		t.Fatalf(
			"expected 2 ledger entries, got %d",
			len(entries),
		)
	}

	rollbackEntry := entries[1]

	if rollbackEntry.Direction != domain.LedgerCredit {
		t.Fatalf(
			"expected ROLLBACK of BET to be CREDIT, got %s",
			rollbackEntry.Direction,
		)
	}
}

func TestProcessWagerTransaction_ConcurrentBets(t *testing.T) {
	ctx := context.Background()

	pool, err := NewPostgresPool()
	if err != nil {
		t.Fatalf("failed to connect to postgres: %v", err)
	}
	defer pool.Close()

	// Limpa dados de execuções anteriores.
	_, err = pool.Exec(ctx, `
	DELETE FROM wallet_ledger
	WHERE wallet_id = 'wallet-concurrency-test'
	`)
	if err != nil {
		t.Fatalf("failed to clean test ledger: %v", err)
	}

	_, err = pool.Exec(ctx, `
	DELETE FROM wager_transactions
	WHERE id IN ('transaction-concurrency-1', 'transaction-concurrency-2')
	`)
	if err != nil {
		t.Fatalf("failed to clean test wager transactions: %v", err)
	}

	_, err = pool.Exec(ctx, `
	DELETE FROM wallets
	WHERE id = 'wallet-concurrency-test'
	`)
	if err != nil {
		t.Fatalf("failed to clean test wallet: %v", err)
	}

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

	now := time.Now().UTC()

	wallet, err := domain.NewWallet(
		"wallet-concurrency-test",
		"player-concurrency-test",
		domain.NewMoney(10000, "BRL"),
		now,
	)
	if err != nil {
		t.Fatalf("failed to create wallet: %v", err)
	}

	err = walletRepository.Create(ctx, wallet)
	if err != nil {
		t.Fatalf("failed to persist wallet: %v", err)
	}
	transaction1, err := domain.NewWagerTransaction(
		"transaction-concurrency-1",
		"provider-1",
		"external-1",
		"idempotency-1",
		"hash-1",
		wallet.PlayerID,
		wallet.ID,
		"round-1",
		"game-1",
		domain.TransactionBet,
		domain.NewMoney(8000, "BRL"),
		"",
		now,
	)
	if err != nil {
		t.Fatalf("failed to create transaction 1: %v", err)
	}

	transaction2, err := domain.NewWagerTransaction(
		"transaction-concurrency-2",
		"provider-1",
		"external-2",
		"idempotency-2",
		"hash-2",
		wallet.PlayerID,
		wallet.ID,
		"round-1",
		"game-1",
		domain.TransactionBet,
		domain.NewMoney(8000, "BRL"),
		"",
		now,
	)
	if err != nil {
		t.Fatalf("failed to create transaction 2: %v", err)
	}

	var wg sync.WaitGroup
	errs := make(chan error, 2)

	wg.Add(2)

	go func() {
		defer wg.Done()
		errs <- processor.Execute(ctx, transaction1)
	}()

	go func() {
		defer wg.Done()
		errs <- processor.Execute(ctx, transaction2)
	}()

	wg.Wait()
	close(errs)

	successCount := 0
	rejectionCount := 0

	for err := range errs {
		if err == nil {
			successCount++
			continue
		}

		if errors.Is(err, ErrWalletConcurrency) ||
			errors.Is(err, domain.ErrInsufficientBalance) {
			rejectionCount++
			continue
		}

		t.Fatalf("unexpected error: %v", err)
	}

	if successCount != 1 {
		t.Fatalf(
			"expected exactly 1 successful transaction, got %d",
			successCount,
		)
	}

	if rejectionCount != 1 {
		t.Fatalf(
			"expected exactly 1 rejected transaction, got %d",
			rejectionCount,
		)
	}

	finalWallet, err := walletRepository.FindByID(ctx, wallet.ID)
	if err != nil {
		t.Fatalf("failed to find final wallet: %v", err)
	}

	if finalWallet.Balance.Amount != 2000 {
		t.Fatalf(
			"expected final balance 2000, got %d",
			finalWallet.Balance.Amount,
		)
	}
}
