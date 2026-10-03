package domain

import (
	"testing"
	"time"
)

func TestNewWagerTransaction(t *testing.T) {
	transaction, err := NewWagerTransaction(
		"tx-123",
		"provider-123",
		"external-tx-123",
		"idempotency-123",
		"hash-123",
		"player-123",
		"wallet-123",
		"round-123",
		"game-123",
		TransactionBet,
		NewMoney(5000, "BRL"),
		"",
		time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC),
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if transaction.ID != "tx-123" {
		t.Fatalf("expected tx-123, got %s", transaction.ID)
	}

	if transaction.WalletID != "wallet-123" {
		t.Fatalf("expected wallet-123, got %s", transaction.WalletID)
	}

	if transaction.RoundID != "round-123" {
		t.Fatalf("expected round-123, got %s", transaction.RoundID)
	}

	if transaction.GameID != "game-123" {
		t.Fatalf("expected game-123, got %s", transaction.GameID)
	}

	if transaction.Type != TransactionBet {
		t.Fatalf("expected BET, got %s", transaction.Type)
	}

	if transaction.Amount.Amount != 5000 {
		t.Fatalf("expected 5000, got %d", transaction.Amount.Amount)
	}

	if transaction.Amount.Currency != "BRL" {
		t.Fatalf("expected BRL, got %s", transaction.Amount.Currency)
	}

	expectedTime := time.Date(
		2026,
		1,
		1,
		12,
		0,
		0,
		0,
		time.UTC,
	)

	if !transaction.CreatedAt.Equal(expectedTime) {
		t.Fatalf("expected CreatedAt %v, got %v", expectedTime, transaction.CreatedAt)
	}

	if !transaction.UpdatedAt.Equal(expectedTime) {
		t.Fatalf("expected UpdatedAt %v, got %v", expectedTime, transaction.UpdatedAt)
	}
}

func TestNewWagerTransactionInvalidType(t *testing.T) {
	_, err := NewWagerTransaction(
		"tx-123",
		"provider-123",
		"external-tx-123",
		"idempotency-123",
		"hash-123",
		"player-123",
		"wallet-123",
		"round-123",
		"game-123",
		TransactionType("INVALID"),
		NewMoney(5000, "BRL"),
		"",
		time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC),
	)

	if err != ErrInvalidTransactionType {
		t.Fatalf("expected invalid transaction type error, got %v", err)
	}

}

func TestTransactionTypeIsValid(t *testing.T) {
	if !TransactionBet.IsValid() {
		t.Fatal("expected BET to be valid")
	}

	if !TransactionWin.IsValid() {
		t.Fatal("expected WIN to be valid")
	}

	if TransactionType("INVALID").IsValid() {
		t.Fatal("expected INVALID to be invalid")
	}
}

func TestWagerTransactionIsBet(t *testing.T) {
	transaction, err := NewWagerTransaction(
		"tx-123",
		"provider-123",
		"external-tx-123",
		"idempotency-123",
		"hash-123",
		"player-123",

		"wallet-123",
		"round-123",
		"game-123",
		TransactionBet,
		NewMoney(5000, "BRL"),
		"",
		time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC),
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !transaction.IsBet() {
		t.Fatal("expected transaction to be a BET")
	}

	if transaction.State != TransactionPending {
		t.Fatalf("expected transaction to start as PENDING, got %s", transaction.State)
	}
}

func TestWagerTransactionStartProcessing(t *testing.T) {
	transaction, err := NewWagerTransaction(
		"tx-123",
		"provider-123",
		"external-tx-123",
		"idempotency-123",
		"hash-123",
		"player-123",

		"wallet-123",
		"round-123",
		"game-123",
		TransactionBet,
		NewMoney(5000, "BRL"),
		"",
		time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC),
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = transaction.StartProcessing()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if transaction.State != TransactionProcessing {
		t.Fatalf(
			"expected PROCESSING, got %s",
			transaction.State,
		)
	}
}

func TestWagerTransactionCannotStartProcessingTwice(t *testing.T) {
	transaction, err := NewWagerTransaction(
		"tx-123",
		"provider-123",
		"external-tx-123",
		"idempotency-123",
		"hash-123",
		"player-123",
		"wallet-123",
		"round-123",
		"game-123",
		TransactionBet,
		NewMoney(5000, "BRL"),
		"",
		time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC),
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = transaction.StartProcessing()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = transaction.StartProcessing()

	if err != ErrInvalidTransactionState {
		t.Fatalf(
			"expected invalid transaction state error, got %v",
			err,
		)
	}
}

func TestWagerTransactionComplete(t *testing.T) {
	transaction, err := NewWagerTransaction(
		"tx-123",
		"provider-123",
		"external-tx-123",
		"idempotency-123",
		"hash-123",
		"player-123",

		"wallet-123",
		"round-123",
		"game-123",
		TransactionBet,
		NewMoney(5000, "BRL"),
		"",
		time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC),
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = transaction.StartProcessing()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = transaction.Complete()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if transaction.State != TransactionCompleted {
		t.Fatalf(
			"expected COMPLETED, got %s",
			transaction.State,
		)
	}
}

func TestWagerTransactionCannotCompleteBeforeProcessing(t *testing.T) {
	transaction, err := NewWagerTransaction(
		"tx-123",
		"provider-123",
		"external-tx-123",
		"idempotency-123",
		"hash-123",
		"player-123",

		"wallet-123",
		"round-123",
		"game-123",
		TransactionBet,
		NewMoney(5000, "BRL"),
		"",
		time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC),
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = transaction.Complete()

	if err != ErrInvalidTransactionState {
		t.Fatalf(
			"expected invalid transaction state error, got %v",
			err,
		)
	}
}

func TestWagerTransactionReject(t *testing.T) {
	transaction, err := NewWagerTransaction(
		"tx-123",
		"provider-123",
		"external-tx-123",
		"idempotency-123",
		"hash-123",
		"player-123",

		"wallet-123",
		"round-123",
		"game-123",
		TransactionBet,
		NewMoney(5000, "BRL"),
		"",
		time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC),
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = transaction.StartProcessing()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = transaction.Reject()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if transaction.State != TransactionRejected {
		t.Fatalf(
			"expected REJECTED, got %s",
			transaction.State,
		)
	}
}

func TestWagerTransactionCannotRejectBeforeProcessing(t *testing.T) {
	transaction, err := NewWagerTransaction(
		"tx-123",
		"provider-123",
		"external-tx-123",
		"idempotency-123",
		"hash-123",
		"player-123",

		"wallet-123",
		"round-123",
		"game-123",
		TransactionBet,
		NewMoney(5000, "BRL"),
		"",
		time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC),
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = transaction.Reject()

	if err != ErrInvalidTransactionState {
		t.Fatalf(
			"expected invalid transaction state error, got %v",
			err,
		)
	}
}

func TestWagerTransactionWaitForReference(t *testing.T) {
	transaction, err := NewWagerTransaction(
		"tx-123",
		"provider-123",
		"external-tx-123",
		"idempotency-123",
		"hash-123",
		"player-123",

		"wallet-123",
		"round-123",
		"game-123",
		TransactionWin,
		NewMoney(5000, "BRL"),
		"bet-123",
		time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC),
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = transaction.StartProcessing()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = transaction.WaitForReference()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if transaction.State != TransactionPendingReference {
		t.Fatalf(
			"expected PENDING_REFERENCE, got %s",
			transaction.State,
		)
	}
}

func TestWagerTransactionResumeProcessing(t *testing.T) {
	transaction, err := NewWagerTransaction(
		"tx-123",
		"provider-123",
		"external-tx-123",
		"idempotency-123",
		"hash-123",
		"player-123",

		"wallet-123",
		"round-123",
		"game-123",
		TransactionWin,
		NewMoney(5000, "BRL"),
		"bet-123",
		time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC),
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = transaction.StartProcessing()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = transaction.WaitForReference()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = transaction.ResumeProcessing()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if transaction.State != TransactionProcessing {
		t.Fatalf(
			"expected PROCESSING, got %s",
			transaction.State,
		)
	}
}

func TestWagerTransactionCannotResumeAfterCompleted(t *testing.T) {
	transaction, err := NewWagerTransaction(
		"tx-123",
		"provider-123",
		"external-tx-123",
		"idempotency-123",
		"hash-123",
		"player-123",

		"wallet-123",
		"round-123",
		"game-123",
		TransactionBet,
		NewMoney(5000, "BRL"),
		"",
		time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatal(err)
	}

	err = transaction.StartProcessing()
	if err != nil {
		t.Fatal(err)
	}

	err = transaction.Complete()
	if err != nil {
		t.Fatal(err)
	}

	err = transaction.ResumeProcessing()
	if err != ErrInvalidTransactionState {
		t.Fatalf("expected ErrInvalidTransactionState, got %v", err)
	}
}

func TestWagerTransactionCannotResumeAfterRejected(t *testing.T) {
	transaction, err := NewWagerTransaction(
		"tx-123",
		"provider-123",
		"external-tx-123",
		"idempotency-123",
		"hash-123",
		"player-123",

		"wallet-123",
		"round-123",
		"game-123",
		TransactionBet,
		NewMoney(5000, "BRL"),
		"",
		time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatal(err)
	}

	err = transaction.StartProcessing()
	if err != nil {
		t.Fatal(err)
	}

	err = transaction.Reject()
	if err != nil {
		t.Fatal(err)
	}

	err = transaction.ResumeProcessing()
	if err != ErrInvalidTransactionState {
		t.Fatalf("expected ErrInvalidTransactionState, got %v", err)
	}
}

func TestWagerTransactionPermanentFail(t *testing.T) {
	transaction, err := NewWagerTransaction(
		"tx-123",
		"provider-123",
		"external-tx-123",
		"idempotency-123",
		"hash-123",
		"player-123",

		"wallet-123",
		"round-123",
		"game-123",
		TransactionBet,
		NewMoney(5000, "BRL"),
		"",
		time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatal(err)
	}

	err = transaction.StartProcessing()
	if err != nil {
		t.Fatal(err)
	}

	err = transaction.PermanentFail()
	if err != nil {
		t.Fatal(err)
	}

	if transaction.State != TransactionPermanentFailed {
		t.Fatalf(
			"expected state %s, got %s",
			TransactionPermanentFailed,
			transaction.State,
		)
	}
}

func TestWagerTransactionCannotPermanentFailBeforeProcessing(t *testing.T) {
	transaction, err := NewWagerTransaction(
		"tx-123",
		"provider-123",
		"external-tx-123",
		"idempotency-123",
		"hash-123",
		"player-123",

		"wallet-123",
		"round-123",
		"game-123",
		TransactionBet,
		NewMoney(5000, "BRL"),
		"",
		time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatal(err)
	}

	err = transaction.PermanentFail()
	if err != ErrInvalidTransactionState {
		t.Fatalf("expected ErrInvalidTransactionState, got %v", err)
	}
}

func TestWagerTransactionCannotResumeAfterPermanentFail(t *testing.T) {
	transaction, err := NewWagerTransaction(
		"tx-123",
		"provider-123",
		"external-tx-123",
		"idempotency-123",
		"hash-123",
		"player-123",

		"wallet-123",
		"round-123",
		"game-123",
		TransactionBet,
		NewMoney(5000, "BRL"),
		"",
		time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatal(err)
	}

	err = transaction.StartProcessing()
	if err != nil {
		t.Fatal(err)
	}

	err = transaction.PermanentFail()
	if err != nil {
		t.Fatal(err)
	}

	err = transaction.ResumeProcessing()
	if err != ErrInvalidTransactionState {
		t.Fatalf("expected ErrInvalidTransactionState, got %v", err)
	}
}

func TestWagerTransactionRequiresExternalTransactionID(t *testing.T) {
	_, err := NewWagerTransaction(
		"tx-123",
		"provider-123",
		"",
		"idempotency-123",
		"hash-123",
		"player-123",

		"wallet-123",
		"round-123",
		"game-123",
		TransactionBet,
		NewMoney(5000, "BRL"),
		"",
		time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC),
	)

	if err != ErrInvalidWagerTransaction {
		t.Fatalf(
			"expected ErrInvalidWagerTransaction, got %v",
			err,
		)
	}
}

func TestWagerTransactionRequiresIdempotencyKey(t *testing.T) {
	_, err := NewWagerTransaction(
		"tx-123",
		"provider-123",
		"external-tx-123",
		"",
		"hash-123",
		"player-123",

		"wallet-123",
		"round-123",
		"game-123",
		TransactionBet,
		NewMoney(5000, "BRL"),
		"",
		time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC),
	)

	if err != ErrInvalidWagerTransaction {
		t.Fatalf(
			"expected ErrInvalidWagerTransaction, got %v",
			err,
		)
	}
}

func TestWagerTransactionRequiresPayloadHash(t *testing.T) {
	_, err := NewWagerTransaction(
		"tx-123",
		"provider-123",
		"external-tx-123",
		"idempotency-123",
		"",
		"player-123",

		"wallet-123",
		"round-123",
		"game-123",
		TransactionBet,
		NewMoney(5000, "BRL"),
		"",
		time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC),
	)

	if err != ErrInvalidWagerTransaction {
		t.Fatalf(
			"expected ErrInvalidWagerTransaction, got %v",
			err,
		)
	}
}

func TestLossMustHaveZeroAmount(t *testing.T) {
	_, err := NewWagerTransaction(
		"tx-123",
		"provider-123",
		"external-tx-123",
		"idempotency-123",
		"hash-123",
		"player-123",

		"wallet-123",
		"round-123",
		"game-123",
		TransactionLoss,
		NewMoney(5000, "BRL"),
		"",
		time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC),
	)

	if err != ErrInvalidTransactionAmount {
		t.Fatalf(
			"expected ErrInvalidTransactionAmount, got %v",
			err,
		)
	}
}

func TestLossCanHaveZeroAmount(t *testing.T) {
	transaction, err := NewWagerTransaction(
		"tx-123",
		"provider-123",
		"external-tx-123",
		"idempotency-123",
		"hash-123",
		"player-123",

		"wallet-123",
		"round-123",
		"game-123",
		TransactionLoss,
		NewMoney(0, "BRL"),
		"",
		time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC),
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if transaction.Amount.Amount != 0 {
		t.Fatalf(
			"expected loss amount to be zero, got %d",
			transaction.Amount.Amount,
		)
	}
}

func TestBetMustHavePositiveAmount(t *testing.T) {
	_, err := NewWagerTransaction(
		"tx-123",
		"provider-123",
		"external-tx-123",
		"idempotency-123",
		"hash-123",
		"player-123",

		"wallet-123",
		"round-123",
		"game-123",
		TransactionBet,
		NewMoney(0, "BRL"),
		"",
		time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC),
	)

	if err != ErrInvalidTransactionAmount {
		t.Fatalf(
			"expected ErrInvalidTransactionAmount, got %v",
			err,
		)
	}
}

func TestBetCanHavePositiveAmount(t *testing.T) {
	transaction, err := NewWagerTransaction(
		"tx-123",
		"provider-123",
		"external-tx-123",
		"idempotency-123",
		"hash-123",
		"player-123",

		"wallet-123",
		"round-123",
		"game-123",
		TransactionBet,
		NewMoney(5000, "BRL"),
		"",
		time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC),
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if transaction.Amount.Amount != 5000 {
		t.Fatalf(
			"expected amount 5000, got %d",
			transaction.Amount.Amount,
		)
	}
}

func TestFinancialTransactionsMustHavePositiveAmount(t *testing.T) {
	transactionTypes := []TransactionType{
		TransactionWin,
		TransactionRefund,
		TransactionRollback,
	}

	for _, transactionType := range transactionTypes {
		t.Run(string(transactionType), func(t *testing.T) {
			_, err := NewWagerTransaction(
				"tx-123",
				"provider-123",
				"external-tx-123",
				"idempotency-123",
				"hash-123",
				"player-123",

				"wallet-123",
				"round-123",
				"game-123",
				transactionType,
				NewMoney(0, "BRL"),
				"",
				time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC),
			)

			if err != ErrInvalidTransactionAmount {
				t.Fatalf(
					"expected ErrInvalidTransactionAmount, got %v",
					err,
				)
			}
		})
	}
}

func TestRefundRequiresReference(t *testing.T) {
	_, err := NewWagerTransaction(
		"tx-123",
		"provider-123",
		"external-tx-123",
		"idempotency-123",
		"hash-123",
		"player-123",

		"wallet-123",
		"round-123",
		"game-123",
		TransactionRefund,
		NewMoney(5000, "BRL"),
		"",
		time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC),
	)

	if err != ErrInvalidWagerTransaction {
		t.Fatalf(
			"expected ErrInvalidWagerTransaction, got %v",
			err,
		)
	}
}

func TestRollbackRequiresReference(t *testing.T) {
	_, err := NewWagerTransaction(
		"tx-123",
		"provider-123",
		"external-tx-123",
		"idempotency-123",
		"hash-123",
		"player-123",

		"wallet-123",
		"round-123",
		"game-123",
		TransactionRollback,
		NewMoney(5000, "BRL"),
		"",
		time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC),
	)

	if err != ErrInvalidWagerTransaction {
		t.Fatalf(
			"expected ErrInvalidWagerTransaction, got %v",
			err,
		)
	}
}

func TestNewOpeningTransaction(t *testing.T) {
	transaction, err := NewOpeningTransaction(
		"opening-123",
		"wallet-123",
		NewMoney(10000, "BRL"),
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if transaction.ID != "opening-123" {
		t.Fatalf("expected ID opening-123, got %s", transaction.ID)
	}

	if transaction.WalletID != "wallet-123" {
		t.Fatalf("expected wallet ID wallet-123, got %s", transaction.WalletID)
	}

	if transaction.Type != TransactionOpening {
		t.Fatalf(
			"expected type %s, got %s",
			TransactionOpening,
			transaction.Type,
		)
	}

	if transaction.State != TransactionPending {
		t.Fatalf(
			"expected state %s, got %s",
			TransactionPending,
			transaction.State,
		)
	}

	if transaction.Amount.Amount != 10000 {
		t.Fatalf(
			"expected amount 10000, got %d",
			transaction.Amount.Amount,
		)
	}
}

func TestNewOpeningTransactionRejectsNegativeAmount(t *testing.T) {
	_, err := NewOpeningTransaction(
		"opening-123",

		"wallet-123",
		NewMoney(-10000, "BRL"),
	)

	if err != ErrInvalidTransactionAmount {
		t.Fatalf(
			"expected ErrInvalidTransactionAmount, got %v",
			err,
		)
	}
}

func TestRehydrateWagerTransaction(t *testing.T) {
	transaction, err := RehydrateWagerTransaction(
		"tx-123",
		"provider-123",
		"external-tx-123",
		"idempotency-123",
		"hash-123",
		"wallet-123",
		"player-123",
		"round-123",
		"game-123",
		TransactionBet,
		NewMoney(5000, "BRL"),
		"",
		TransactionCompleted,
		time.Now(),
		time.Now(),
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if transaction.ID != "tx-123" {
		t.Fatalf("expected ID tx-123, got %s", transaction.ID)
	}

	if transaction.State != TransactionCompleted {
		t.Fatalf(
			"expected state %s, got %s",
			TransactionCompleted,
			transaction.State,
		)
	}
}

func TestRehydrateWagerTransactionRejectsInvalidState(t *testing.T) {
	_, err := RehydrateWagerTransaction(
		"tx-123",
		"provider-123",
		"external-tx-123",
		"idempotency-123",
		"hash-123",
		"wallet-123",
		"player-123",
		"round-123",
		"game-123",
		TransactionBet,
		NewMoney(5000, "BRL"),
		"",
		TransactionState("INVALID"),
		time.Now(),
		time.Now(),
	)

	if err != ErrInvalidTransactionState {
		t.Fatalf(
			"expected ErrInvalidTransactionState, got %v",
			err,
		)
	}
}

func TestRefundCannotReferenceItself(t *testing.T) {
	_, err := NewWagerTransaction(
		"tx-123",
		"provider-123",
		"external-tx-123",
		"idempotency-123",
		"hash-123",
		"player-123",

		"wallet-123",
		"round-123",
		"game-123",
		TransactionRefund,
		NewMoney(5000, "BRL"),
		"external-tx-123",
		time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC),
	)

	if err != ErrInvalidWagerTransaction {
		t.Fatalf(
			"expected ErrInvalidWagerTransaction, got %v",
			err,
		)
	}
}

func TestRollbackCannotReferenceItself(t *testing.T) {
	_, err := NewWagerTransaction(
		"tx-123",
		"provider-123",
		"external-tx-123",
		"idempotency-123",
		"hash-123",
		"player-123",

		"wallet-123",
		"round-123",
		"game-123",
		TransactionRollback,
		NewMoney(5000, "BRL"),
		"external-tx-123",
		time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC),
	)

	if err != ErrInvalidWagerTransaction {
		t.Fatalf(
			"expected ErrInvalidWagerTransaction, got %v",
			err,
		)
	}
}

func TestNewWagerTransactionRejectsOpening(t *testing.T) {
	_, err := NewWagerTransaction(
		"tx-1",
		"provider-1",
		"external-1",
		"idem-1",
		"hash-1",
		"player-1",
		"wallet-1",
		"round-1",
		"game-1",
		TransactionOpening,
		NewMoney(1000, "BRL"),
		"",
		time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC),
	)

	if err != ErrInvalidTransactionType {
		t.Fatalf("expected ErrInvalidTransactionType, got %v", err)
	}
}

func TestWagerTransactionRequiresRoundID(t *testing.T) {
	_, err := NewWagerTransaction(
		"tx-123",
		"provider-123",
		"external-tx-123",
		"idempotency-123",
		"hash-123",
		"player-123",

		"wallet-123",
		"",
		"game-1",
		TransactionBet,
		NewMoney(5000, "BRL"),
		"",
		time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC),
	)

	if err != ErrInvalidWagerTransaction {
		t.Fatalf(
			"expected ErrInvalidWagerTransaction, got %v",
			err,
		)
	}
}

func TestWagerTransactionRequiresGameID(t *testing.T) {
	_, err := NewWagerTransaction(
		"tx-123",
		"provider-123",
		"external-tx-123",
		"idempotency-123",
		"hash-123",
		"player-123",

		"wallet-123",
		"round-123",
		"",
		TransactionBet,
		NewMoney(5000, "BRL"),
		"",
		time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC),
	)

	if err != ErrInvalidWagerTransaction {
		t.Fatalf(
			"expected ErrInvalidWagerTransaction, got %v",
			err,
		)
	}
}

func TestWagerTransactionRequiresPlayerID(t *testing.T) {
	_, err := NewWagerTransaction(
		"tx-123",
		"provider-123",
		"external-tx-123",
		"idempotency-123",
		"hash-123",
		"",
		"wallet-123",
		"round-123",
		"game-123",
		TransactionBet,
		NewMoney(5000, "BRL"),
		"",
		time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC),
	)

	if err != ErrInvalidWagerTransaction {
		t.Fatalf(
			"expected ErrInvalidWagerTransaction, got %v",
			err,
		)
	}
}
