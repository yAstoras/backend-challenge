package domain

import (
	"errors"
	"time"
)

type TransactionType string

type TransactionState string

const (
	TransactionPending          TransactionState = "PENDING"
	TransactionProcessing       TransactionState = "PROCESSING"
	TransactionPendingReference TransactionState = "PENDING_REFERENCE"
	TransactionCompleted        TransactionState = "COMPLETED"
	TransactionRejected         TransactionState = "REJECTED"
	TransactionFailed           TransactionState = "FAILED"
	TransactionProcessed        TransactionState = "PROCESSED"
	TransactionPermanentFailed  TransactionState = "PERMANENT_FAILED"
)

var (
	ErrInvalidTransactionType   = errors.New("invalid transaction type")
	ErrInvalidWagerTransaction  = errors.New("invalid wager transaction")
	ErrInvalidTransactionState  = errors.New("invalid transaction state")
	ErrInvalidTransactionAmount = errors.New("invalid transaction amount")
)

const (
	TransactionOpening  TransactionType = "OPENING"
	TransactionBet      TransactionType = "BET"
	TransactionWin      TransactionType = "WIN"
	TransactionLoss     TransactionType = "LOSS"
	TransactionRefund   TransactionType = "REFUND"
	TransactionRollback TransactionType = "ROLLBACK"
)

type WagerTransaction struct {
	PlayerID                       string
	ID                             string
	FailureCode                    string
	ProviderID                     string
	ExternalTransactionID          string
	IdempotencyKey                 string
	PayloadHash                    string
	WalletID                       string
	RoundID                        string
	GameID                         string
	Type                           TransactionType
	Amount                         Money
	ReferenceExternalTransactionID string
	State                          TransactionState
	CreatedAt                      time.Time
	UpdatedAt                      time.Time
	ReferenceAttempts              int
	NextReferenceAttemptAt         time.Time
}

func NewWagerTransaction(
	id string,
	providerID string,
	externalTransactionID string,
	idempotencyKey string,
	payloadHash string,
	playerID string,
	walletID string,
	roundID string,
	gameID string,
	transactionType TransactionType,
	amount Money,
	referenceExternalTransactionID string,
	now time.Time,
) (WagerTransaction, error) {
	if playerID == "" {
		return WagerTransaction{}, ErrInvalidWagerTransaction
	}

	if id == "" {
		return WagerTransaction{}, ErrInvalidWagerTransaction
	}

	if providerID == "" {
		return WagerTransaction{}, ErrInvalidWagerTransaction
	}

	if externalTransactionID == "" {
		return WagerTransaction{}, ErrInvalidWagerTransaction
	}

	if idempotencyKey == "" {
		return WagerTransaction{}, ErrInvalidWagerTransaction
	}

	if payloadHash == "" {
		return WagerTransaction{}, ErrInvalidWagerTransaction
	}

	if walletID == "" {
		return WagerTransaction{}, ErrInvalidWagerTransaction
	}

	if roundID == "" {
		return WagerTransaction{}, ErrInvalidWagerTransaction
	}

	if gameID == "" {
		return WagerTransaction{}, ErrInvalidWagerTransaction
	}

	if !transactionType.IsValid() {
		return WagerTransaction{}, ErrInvalidTransactionType
	}

	if transactionType == TransactionOpening {
		return WagerTransaction{}, ErrInvalidTransactionType
	}

	if transactionType == TransactionLoss && amount.Amount != 0 {
		return WagerTransaction{}, ErrInvalidTransactionAmount
	}

	if transactionType == TransactionBet && amount.Amount <= 0 {
		return WagerTransaction{}, ErrInvalidTransactionAmount
	}

	if (transactionType == TransactionWin ||
		transactionType == TransactionRefund ||
		transactionType == TransactionRollback) &&
		amount.Amount <= 0 {
		return WagerTransaction{}, ErrInvalidTransactionAmount
	}

	if (transactionType == TransactionRefund ||
		transactionType == TransactionRollback) &&
		referenceExternalTransactionID == "" {
		return WagerTransaction{}, ErrInvalidWagerTransaction
	}

	if (transactionType == TransactionRefund ||
		transactionType == TransactionRollback) &&
		referenceExternalTransactionID == externalTransactionID {
		return WagerTransaction{}, ErrInvalidWagerTransaction
	}

	if now.IsZero() {
		return WagerTransaction{}, ErrInvalidWagerTransaction
	}

	return WagerTransaction{
		PlayerID:                       playerID,
		ID:                             id,
		ProviderID:                     providerID,
		ExternalTransactionID:          externalTransactionID,
		IdempotencyKey:                 idempotencyKey,
		PayloadHash:                    payloadHash,
		WalletID:                       walletID,
		RoundID:                        roundID,
		GameID:                         gameID,
		Type:                           transactionType,
		Amount:                         amount,
		ReferenceExternalTransactionID: referenceExternalTransactionID,
		State:                          TransactionPending,
		CreatedAt:                      now,
		UpdatedAt:                      now,
	}, nil
}

func (w *WagerTransaction) StartProcessing() error {
	if w.State != TransactionPending {
		return ErrInvalidTransactionState
	}

	w.State = TransactionProcessing

	return nil
}

func (w *WagerTransaction) Complete() error {
	if w.State != TransactionProcessing {
		return ErrInvalidTransactionState
	}

	w.State = TransactionCompleted

	return nil
}

func (w *WagerTransaction) Reject() error {
	if w.State != TransactionProcessing {
		return ErrInvalidTransactionState
	}

	w.State = TransactionRejected

	return nil
}

func (w *WagerTransaction) Fail() error {
	if w.State != TransactionProcessing {
		return ErrInvalidTransactionState
	}

	w.State = TransactionFailed

	return nil
}

func (w *WagerTransaction) Process() error {
	if w.State != TransactionProcessing {
		return ErrInvalidTransactionState
	}

	w.State = TransactionProcessed

	return nil
}

func (w *WagerTransaction) WaitForReference() error {
	if w.State != TransactionProcessing {
		return ErrInvalidTransactionState
	}

	w.State = TransactionPendingReference

	return nil
}

func (w *WagerTransaction) ResumeProcessing() error {
	if w.State != TransactionPendingReference {
		return ErrInvalidTransactionState
	}

	w.State = TransactionProcessing

	return nil
}

func (t TransactionType) IsValid() bool {
	switch t {
	case TransactionOpening,
		TransactionBet,
		TransactionWin,
		TransactionLoss,
		TransactionRefund,
		TransactionRollback:
		return true
	default:
		return false
	}
}

func (w *WagerTransaction) PermanentFail() error {
	if w.State != TransactionProcessing {
		return ErrInvalidTransactionState
	}

	w.State = TransactionPermanentFailed

	return nil
}

func (w WagerTransaction) IsBet() bool {
	return w.Type == TransactionBet
}

func NewOpeningTransaction(
	id string,
	walletID string,
	amount Money,
) (WagerTransaction, error) {
	if id == "" {
		return WagerTransaction{}, ErrInvalidWagerTransaction
	}

	if walletID == "" {
		return WagerTransaction{}, ErrInvalidWagerTransaction
	}

	if amount.Amount < 0 {
		return WagerTransaction{}, ErrInvalidTransactionAmount
	}

	return WagerTransaction{
		ID:       id,
		WalletID: walletID,
		Type:     TransactionOpening,
		Amount:   amount,
		State:    TransactionPending,
	}, nil
}

func RehydrateWagerTransaction(
	id string,
	providerID string,
	externalTransactionID string,
	idempotencyKey string,
	payloadHash string,
	walletID string,
	playerID string,
	roundID string,
	gameID string,
	transactionType TransactionType,
	amount Money,
	referenceExternalTransactionID string,
	state TransactionState,
	createdAt time.Time,
	updatedAt time.Time,
) (WagerTransaction, error) {
	if id == "" {
		return WagerTransaction{}, ErrInvalidWagerTransaction
	}

	if walletID == "" {
		return WagerTransaction{}, ErrInvalidWagerTransaction
	}

	if !transactionType.IsValid() {
		return WagerTransaction{}, ErrInvalidTransactionType
	}

	if playerID == "" || roundID == "" || gameID == "" {
		return WagerTransaction{}, ErrInvalidWagerTransaction
	}

	if createdAt.IsZero() || updatedAt.IsZero() {
		return WagerTransaction{}, ErrInvalidWagerTransaction
	}

	if updatedAt.Before(createdAt) {
		return WagerTransaction{}, ErrInvalidWagerTransaction
	}

	switch state {
	case TransactionPending,
		TransactionProcessing,
		TransactionPendingReference,
		TransactionCompleted,
		TransactionRejected,
		TransactionFailed,
		TransactionProcessed,
		TransactionPermanentFailed:
		// estado válido
	default:
		return WagerTransaction{}, ErrInvalidTransactionState
	}

	return WagerTransaction{
		ID:                             id,
		ProviderID:                     providerID,
		ExternalTransactionID:          externalTransactionID,
		IdempotencyKey:                 idempotencyKey,
		PayloadHash:                    payloadHash,
		WalletID:                       walletID,
		Type:                           transactionType,
		Amount:                         amount,
		ReferenceExternalTransactionID: referenceExternalTransactionID,
		State:                          state,
		PlayerID:                       playerID,
		RoundID:                        roundID,
		GameID:                         gameID,
		CreatedAt:                      createdAt,
		UpdatedAt:                      updatedAt,
	}, nil
}
