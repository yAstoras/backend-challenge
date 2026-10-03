package application

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yasto/backend-challenge/internal/domain"
)

type SQSConsumer struct {
	client          *SQSClient
	queueURL        string
	db              *pgxpool.Pool
	inboxRepository *InboxRepository
	processor       *PostgresWagerTransactionProcessor
}

func NewSQSConsumer(
	client *SQSClient,
	queueURL string,
	db *pgxpool.Pool,
	inboxRepository *InboxRepository,
	processor *PostgresWagerTransactionProcessor,
) *SQSConsumer {
	return &SQSConsumer{
		client:          client,
		queueURL:        queueURL,
		db:              db,
		inboxRepository: inboxRepository,
		processor:       processor,
	}
}

func (c *SQSConsumer) Receive(ctx context.Context) (WagerTransactionMessage, string, string, string, error) {
	messageID, body, receiptHandle, err := c.client.ReceiveMessage(ctx, c.queueURL)
	if err != nil {
		return WagerTransactionMessage{}, "", "", "", err
	}

	var message WagerTransactionMessage

	if err := json.Unmarshal([]byte(body), &message); err != nil {
		return WagerTransactionMessage{}, "", "", "", fmt.Errorf("decode SQS message: %w", err)
	}

	if messageID == "" {
		return WagerTransactionMessage{}, "", "", "", fmt.Errorf("SQS message has no message ID")
	}

	if message.Data.IdempotencyKey == "" {
		return WagerTransactionMessage{}, "", "", "", fmt.Errorf("SQS message has no idempotency key")
	}

	return message, messageID, receiptHandle, body, nil
}

func buildWagerTransactionFromMessage(
	message WagerTransactionMessage,
	payloadHash string,
	now time.Time,
) (domain.WagerTransaction, error) {
	transactionType := domain.TransactionType(message.Data.Type)

	money := domain.NewMoney(
		message.Data.Amount,
		message.Data.Currency,
	)

	return domain.NewWagerTransaction(
		message.Data.TransactionID,
		message.Data.ProviderID,
		message.Data.ExternalTransactionID,
		message.Data.IdempotencyKey,
		payloadHash,
		message.Data.PlayerID,
		message.Data.WalletID,
		message.Data.RoundID,
		message.Data.GameID,
		transactionType,
		money,
		message.Data.ReferenceExternalTransactionID,
		now,
	)
}

func hashPayload(body string) string {
	hash := sha256.Sum256([]byte(body))
	return hex.EncodeToString(hash[:])
}

func (c *SQSConsumer) ProcessMessage(ctx context.Context) error {
	message, messageID, receiptHandle, body, err := c.Receive(ctx)
	if err != nil {
		return err
	}

	payloadHash := hashPayload(body)

	transaction, err := buildWagerTransactionFromMessage(
		message,
		payloadHash,
		time.Now(),
	)
	if err != nil {
		return err
	}

	tx, err := c.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin SQS transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	created, err := c.inboxRepository.CreateTx(
		ctx,
		tx,
		messageID,
		message.Data.IdempotencyKey,
	)
	if err != nil {
		return err
	}

	if !created {
		if err := tx.Commit(ctx); err != nil {
			return err
		}

		return c.client.DeleteMessage(ctx, c.queueURL, receiptHandle)
	}

	err = c.processor.executeTx(ctx, tx, transaction)

	if err != nil && !errors.Is(err, ErrTransactionRejected) {
		return err
	}

	if err := c.inboxRepository.MarkProcessedTx(
		ctx,
		tx,
		messageID,
	); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit SQS transaction: %w", err)
	}

	if err := c.client.DeleteMessage(ctx, c.queueURL, receiptHandle); err != nil {
		return err
	}

	return err
}

func (c *SQSConsumer) Start(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return

		case <-ticker.C:
			err := c.ProcessMessage(ctx)
			if err != nil {
				fmt.Printf("SQS consumer error: %v\n", err)
			}
		}
	}
}
