package application

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yasto/backend-challenge/internal/domain"
)

var (
	ErrReferenceNotFound   = errors.New("reference transaction not found")
	ErrInvalidReference    = errors.New("invalid reference transaction")
	ErrDuplicateReversal   = errors.New("duplicate reversal")
	ErrReferenceMismatch   = errors.New("reference transaction mismatch")
	ErrTransactionRejected = errors.New("transaction rejected")
)

const (
	FailureCodeInsufficientBalance         = "INSUFFICIENT_BALANCE"
	FailureCodeInvalidReference            = "INVALID_REFERENCE"
	FailureCodeReferenceMismatch           = "REFERENCE_MISMATCH"
	FailureCodeDuplicateReversal           = "DUPLICATE_REVERSAL"
	FailureCodeReversalInsufficientBalance = "REVERSAL_INSUFFICIENT_BALANCE"
	FailureCodeReferenceNotFound           = "REFERENCE_NOT_FOUND"
)

const (
	referenceRetryBaseDelay   = time.Second
	referenceRetryMaxAttempts = 5
)

func nextReferenceRetryAt(attempts int, now time.Time) time.Time {
	delay := referenceRetryBaseDelay

	for i := 1; i < attempts; i++ {
		delay *= 2
	}

	return now.Add(delay)
}

type transactionRejectedError struct {
	cause error
}

func (e transactionRejectedError) Error() string {
	return ErrTransactionRejected.Error()
}

func (e transactionRejectedError) Is(target error) bool {
	return target == ErrTransactionRejected || errors.Is(e.cause, target)
}

type PostgresWagerTransactionProcessor struct {
	db                         *pgxpool.Pool
	walletRepository           *PostgresWalletRepository
	ledgerRepository           *PostgresLedgerRepository
	wagerTransactionRepository *PostgresWagerTransactionRepository
	outboxRepository           *OutboxRepository
}

func NewPostgresWagerTransactionProcessor(
	db *pgxpool.Pool,
	walletRepository *PostgresWalletRepository,
	ledgerRepository *PostgresLedgerRepository,
	wagerTransactionRepository *PostgresWagerTransactionRepository,
	outboxRepository *OutboxRepository,
) *PostgresWagerTransactionProcessor {
	return &PostgresWagerTransactionProcessor{
		db:                         db,
		walletRepository:           walletRepository,
		ledgerRepository:           ledgerRepository,
		wagerTransactionRepository: wagerTransactionRepository,
		outboxRepository:           outboxRepository,
	}
}

func (p *PostgresWagerTransactionProcessor) Execute(
	ctx context.Context,
	transaction domain.WagerTransaction,
) error {
	existing, err := p.wagerTransactionRepository.FindByIdempotencyKey(
		ctx,
		transaction.IdempotencyKey,
	)
	if err == nil {
		if existing.PayloadHash != transaction.PayloadHash {
			return ErrIdempotencyConflict
		}

		return nil
	}

	if !errors.Is(err, ErrWagerTransactionNotFound) {
		return err
	}

	existing, err = p.wagerTransactionRepository.FindByProviderExternalID(
		ctx,
		transaction.ProviderID,
		transaction.ExternalTransactionID,
	)
	if err == nil {
		return ErrProviderTransactionConflict
	}

	if !errors.Is(err, ErrWagerTransactionNotFound) {
		return err
	}

	tx, err := p.db.Begin(ctx)
	if err != nil {
		return err
	}

	err = p.executeTx(ctx, tx, transaction)
	if err != nil && !errors.Is(err, ErrTransactionRejected) {
		_ = tx.Rollback(ctx)

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			existing, findErr := p.wagerTransactionRepository.FindByIdempotencyKey(
				ctx,
				transaction.IdempotencyKey,
			)

			if findErr == nil {
				if existing.PayloadHash != transaction.PayloadHash {
					return ErrIdempotencyConflict
				}

				return nil
			}

			if !errors.Is(findErr, ErrWagerTransactionNotFound) {
				return findErr
			}

			return ErrProviderTransactionConflict
		}

		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	return err
}

func (p *PostgresWagerTransactionProcessor) executeTx(
	ctx context.Context,
	tx pgx.Tx,
	transaction domain.WagerTransaction,
) error {
	if err := p.wagerTransactionRepository.CreateTx(ctx, tx, transaction); err != nil {
		return err
	}

	if err := transaction.StartProcessing(); err != nil {
		return err
	}

	if err := p.processTransaction(ctx, tx, &transaction); err != nil {
		if errors.Is(err, ErrTransactionRejected) {
			if err := p.wagerTransactionRepository.UpdateTx(
				ctx,
				tx,
				transaction,
			); err != nil {
				return err
			}

			event := OutboxEvent{
				ID:            transaction.ID + ":rejected",
				EventType:     "WagerTransactionRejected",
				AggregateID:   transaction.ID,
				CorrelationID: transaction.ID,
				OccurredAt:    time.Now(),
				Version:       1,
				Payload: map[string]any{
					"transactionId": transaction.ID,
					"failureCode":   transaction.FailureCode,
				},
				Attempts:      0,
				NextAttemptAt: time.Now(),
			}

			if err := p.outboxRepository.CreateTx(ctx, tx, event); err != nil {
				return err
			}

			return err
		}

		return err
	}

	if transaction.State == domain.TransactionPendingReference {
		event := OutboxEvent{
			ID:            transaction.ID + ":pending-reference",
			EventType:     "WagerTransactionPendingReference",
			AggregateID:   transaction.ID,
			CorrelationID: transaction.ID,
			OccurredAt:    time.Now(),
			Version:       1,
			Payload: map[string]any{
				"transactionId": transaction.ID,
				"walletId":      transaction.WalletID,
				"attempts":      transaction.ReferenceAttempts,
				"nextAttemptAt": transaction.NextReferenceAttemptAt,
			},
			Attempts:      0,
			NextAttemptAt: time.Now(),
		}

		if err := p.outboxRepository.CreateTx(ctx, tx, event); err != nil {
			return err
		}
	}

	return nil
}

func (p *PostgresWagerTransactionProcessor) ResumePendingReference(
	ctx context.Context,
	transactionID string,
) error {
	tx, err := p.db.Begin(ctx)
	if err != nil {
		return err
	}

	defer func() {
		_ = tx.Rollback(ctx)
	}()

	transaction, err := p.wagerTransactionRepository.FindByIDTx(
		ctx,
		tx,
		transactionID,
	)
	if err != nil {
		return err
	}

	if err := transaction.ResumeProcessing(); err != nil {
		return err
	}

	if err := p.processTransaction(
		ctx,
		tx,
		&transaction,
	); err != nil {
		if errors.Is(err, ErrTransactionRejected) {
			if err := p.wagerTransactionRepository.UpdateTx(
				ctx,
				tx,
				transaction,
			); err != nil {
				return err
			}

			event := OutboxEvent{
				ID:            transaction.ID + ":rejected",
				EventType:     "WagerTransactionRejected",
				AggregateID:   transaction.ID,
				CorrelationID: transaction.ID,
				OccurredAt:    time.Now(),
				Version:       1,
				Payload: map[string]any{
					"transactionId": transaction.ID,
					"failureCode":   transaction.FailureCode,
				},
				Attempts:      0,
				NextAttemptAt: time.Now(),
			}

			if err := p.outboxRepository.CreateTx(ctx, tx, event); err != nil {
				return err
			}

			if err := tx.Commit(ctx); err != nil {
				return err
			}

			return err
		}

		return err
	}

	return tx.Commit(ctx)
}

func (p *PostgresWagerTransactionProcessor) ResumePendingReferences(
	ctx context.Context,
	limit int,
) error {
	transactions, err := p.wagerTransactionRepository.FindPendingReference(
		ctx,
		limit,
	)
	if err != nil {
		return err
	}

	for _, transaction := range transactions {
		if err := p.ResumePendingReference(ctx, transaction.ID); err != nil {
			return err
		}
	}

	return nil
}

func (p *PostgresWagerTransactionProcessor) processTransaction(
	ctx context.Context,
	tx pgx.Tx,
	transaction *domain.WagerTransaction,
) error {
	wallet, err := p.walletRepository.FindByIDForUpdate(
		ctx,
		tx,
		transaction.WalletID,
	)
	if err != nil {
		return err
	}

	balanceBefore := wallet.Balance
	versionBefore := wallet.Version

	switch transaction.Type {
	case domain.TransactionBet:
		err = wallet.Debit(transaction.Amount)

		if errors.Is(err, domain.ErrInsufficientBalance) {
			return rejectTransaction(
				transaction,
				FailureCodeInsufficientBalance,
				err,
			)
		}

	case domain.TransactionWin:
		err = wallet.Credit(transaction.Amount)

	case domain.TransactionLoss:
		// LOSS não altera saldo.

	case domain.TransactionRefund:
		reference, err := p.wagerTransactionRepository.FindByProviderExternalIDTx(
			ctx,
			tx,
			transaction.ProviderID,
			transaction.ReferenceExternalTransactionID,
		)
		if errors.Is(err, ErrWagerTransactionNotFound) {
			transaction.ReferenceAttempts++

			if transaction.ReferenceAttempts >= referenceRetryMaxAttempts {
				return rejectTransaction(
					transaction,
					FailureCodeReferenceNotFound,
					ErrReferenceNotFound,
				)
			}

			if err := transaction.WaitForReference(); err != nil {
				return err
			}

			transaction.NextReferenceAttemptAt = nextReferenceRetryAt(
				transaction.ReferenceAttempts,
				time.Now(),
			)

			if err := p.wagerTransactionRepository.UpdateTx(
				ctx,
				tx,
				*transaction,
			); err != nil {
				return err
			}

			return nil
		}

		if reference.Type != domain.TransactionBet {
			return rejectTransaction(
				transaction,
				FailureCodeInvalidReference,
				ErrInvalidReference,
			)
		}

		if reference.State != domain.TransactionProcessed {
			return rejectTransaction(
				transaction,
				FailureCodeInvalidReference,
				ErrInvalidReference,
			)
		}

		if reference.ProviderID != transaction.ProviderID ||
			reference.PlayerID != transaction.PlayerID ||
			reference.WalletID != transaction.WalletID ||
			reference.RoundID != transaction.RoundID ||
			reference.GameID != transaction.GameID ||
			reference.Amount != transaction.Amount ||
			reference.Amount.Currency != transaction.Amount.Currency {
			return rejectTransaction(
				transaction,
				FailureCodeReferenceMismatch,
				ErrReferenceMismatch,
			)
		}

		_, reversalErr := p.wagerTransactionRepository.FindProcessedReversalByReferenceTx(
			ctx,
			tx,
			transaction.ProviderID,
			transaction.ReferenceExternalTransactionID,
		)

		if reversalErr == nil {
			return rejectTransaction(
				transaction,
				FailureCodeDuplicateReversal,
				ErrDuplicateReversal,
			)
		}

		if !errors.Is(reversalErr, ErrWagerTransactionNotFound) {
			return reversalErr
		}

		err = wallet.Credit(transaction.Amount)

	case domain.TransactionRollback:
		reference, err := p.wagerTransactionRepository.FindByProviderExternalIDTx(
			ctx,
			tx,
			transaction.ProviderID,
			transaction.ReferenceExternalTransactionID,
		)
		if err != nil {
			if errors.Is(err, ErrWagerTransactionNotFound) {
				transaction.ReferenceAttempts++

				if transaction.ReferenceAttempts >= referenceRetryMaxAttempts {
					return rejectTransaction(
						transaction,
						FailureCodeReferenceNotFound,
						ErrReferenceNotFound,
					)
				}

				if err := transaction.WaitForReference(); err != nil {
					return err
				}

				transaction.NextReferenceAttemptAt = nextReferenceRetryAt(
					transaction.ReferenceAttempts,
					time.Now(),
				)

				if err := p.wagerTransactionRepository.UpdateTx(
					ctx,
					tx,
					*transaction,
				); err != nil {
					return err
				}

				return nil
			}
		}

		if reference.State != domain.TransactionProcessed {
			return rejectTransaction(
				transaction,
				FailureCodeInvalidReference,
				ErrInvalidReference,
			)
		}

		if reference.Type != domain.TransactionBet &&
			reference.Type != domain.TransactionWin &&
			reference.Type != domain.TransactionRefund {
			return rejectTransaction(
				transaction,
				FailureCodeInvalidReference,
				ErrInvalidReference,
			)
		}

		if reference.PlayerID != transaction.PlayerID ||
			reference.WalletID != transaction.WalletID ||
			reference.RoundID != transaction.RoundID ||
			reference.ProviderID != transaction.ProviderID ||
			reference.GameID != transaction.GameID ||
			reference.Amount != transaction.Amount ||
			reference.Amount.Currency != transaction.Amount.Currency {
			return rejectTransaction(
				transaction,
				FailureCodeReferenceMismatch,
				ErrReferenceMismatch,
			)
		}

		_, reversalErr := p.wagerTransactionRepository.FindProcessedReversalByReferenceTx(
			ctx,
			tx,
			transaction.ProviderID,
			transaction.ReferenceExternalTransactionID,
		)

		if reversalErr == nil {
			return rejectTransaction(
				transaction,
				FailureCodeDuplicateReversal,
				ErrDuplicateReversal,
			)
		}

		if !errors.Is(reversalErr, ErrWagerTransactionNotFound) {
			return reversalErr
		}

		switch reference.Type {

		case domain.TransactionBet:
			// BET retirou dinheiro; ROLLBACK devolve.
			err = wallet.Credit(transaction.Amount)

		case domain.TransactionWin, domain.TransactionRefund:
			// WIN/REFUND adicionaram dinheiro; ROLLBACK retira.
			err = wallet.Debit(transaction.Amount)
		}

		if errors.Is(err, domain.ErrInsufficientBalance) {
			return rejectTransaction(
				transaction,
				FailureCodeReversalInsufficientBalance,
				err,
			)
		}

	default:
		return ErrUnsupportedTransactionType
	}

	if err != nil {
		return err
	}

	if err := transaction.Process(); err != nil {
		return err
	}

	if err := p.wagerTransactionRepository.UpdateTx(
		ctx,
		tx,
		*transaction,
	); err != nil {
		return err
	}

	direction := "credit"

	if transaction.Type == domain.TransactionBet ||
		transaction.Type == domain.TransactionWin ||
		transaction.Type == domain.TransactionRefund ||
		transaction.Type == domain.TransactionRollback {

		direction := domain.LedgerDebit

		switch transaction.Type {
		case domain.TransactionWin, domain.TransactionRefund:
			direction = domain.LedgerCredit

		case domain.TransactionRollback:
			reference, err := p.wagerTransactionRepository.FindByProviderExternalIDTx(
				ctx,
				tx,
				transaction.ProviderID,
				transaction.ReferenceExternalTransactionID,
			)
			if err != nil {
				return err
			}

			if reference.Type == domain.TransactionBet {
				direction = domain.LedgerCredit
			} else {
				direction = domain.LedgerDebit
			}
		}
		ledgerEntry, err := domain.NewWalletLedgerEntry(
			transaction.ID+"-ledger",
			wallet.ID,
			transaction.ID,
			direction,
			transaction.Amount,
			balanceBefore,
			wallet.Balance,
			time.Now(),
		)
		if err != nil {
			return err
		}

		if err := p.ledgerRepository.SaveTx(ctx, tx, ledgerEntry); err != nil {
			return err
		}
	}

	if wallet.Version != versionBefore {
		if err := p.walletRepository.SaveTx(
			ctx,
			tx,
			wallet,
		); err != nil {
			return err
		}
	}

	event := OutboxEvent{
		ID:            transaction.ID + ":processed",
		EventType:     "WagerTransactionProcessed",
		AggregateID:   transaction.ID,
		CorrelationID: transaction.ID,
		OccurredAt:    time.Now(),
		Version:       1,
		Payload: map[string]any{
			"transactionId": transaction.ID,
			"walletId":      wallet.ID,
			"amount":        transaction.Amount,
			"state":         transaction.State,
		},
		Attempts:      0,
		NextAttemptAt: time.Now(),
	}

	if err := p.outboxRepository.CreateTx(ctx, tx, event); err != nil {
		return err
	}

	if wallet.Version != versionBefore {
		event := OutboxEvent{
			ID:            transaction.ID + ":balance-changed",
			EventType:     "WalletBalanceChanged",
			AggregateID:   wallet.ID,
			CorrelationID: transaction.ID,
			OccurredAt:    time.Now(),
			Version:       1,
			Payload: map[string]any{
				"walletId":      wallet.ID,
				"transactionId": transaction.ID,
				"direction":     direction,
				"money":         transaction.Amount,
				"balanceBefore": balanceBefore,
				"balanceAfter":  wallet.Balance,
				"walletVersion": wallet.Version,
			},
			Attempts:      0,
			NextAttemptAt: time.Now(),
		}

		if err := p.outboxRepository.CreateTx(ctx, tx, event); err != nil {
			return err
		}
	}

	return nil
}

func rejectTransaction(
	transaction *domain.WagerTransaction,
	failureCode string,
	cause error,
) error {
	transaction.FailureCode = failureCode

	if err := transaction.Reject(); err != nil {
		return err
	}

	return transactionRejectedError{
		cause: cause,
	}
}

func (p *PostgresWagerTransactionProcessor) StartReferenceWorker(
	ctx context.Context,
	interval time.Duration,
) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return

		case <-ticker.C:
			_ = p.ResumePendingReferences(ctx, 100)
		}
	}
}
