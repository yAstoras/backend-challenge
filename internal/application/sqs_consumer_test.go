package application

import (
	"context"
	"testing"
	"time"

	"github.com/yasto/backend-challenge/internal/domain"
)

func TestSQSConsumerReceive(t *testing.T) {
	ctx := context.Background()

	client, err := NewSQSClient()
	if err != nil {
		t.Fatalf("failed to create SQS client: %v", err)
	}

	queueURL, err := client.QueueURL(ctx, "wager-transactions")
	if err != nil {
		t.Fatalf("failed to get queue URL: %v", err)
	}

	body := `{
		"messageId": "body-message-id",
		"correlationId": "correlation-001",
		"data": {
			"idempotencyKey": "consumer-test-001",
			"transactionId": "transaction-001",
			"playerId": "player-001",
			"providerId": "provider-001",
			"externalTransactionId": "external-001",
			"walletId": "wallet-001",
			"roundId": "round-001",
			"gameId": "game-001",
			"type": "BET",
			"amount": 1000,
			"currency": "BRL"
		}
	}`

	if err := client.SendMessage(ctx, queueURL, body); err != nil {
		t.Fatalf("failed to send message: %v", err)
	}

	consumer := NewSQSConsumer(
		client,
		queueURL,
		nil,
		nil,
		nil,
	)

	message, messageID, receiptHandle, body, err := consumer.Receive(ctx)
	if err != nil {
		t.Fatalf("failed to receive message: %v", err)
	}

	if messageID == "" {
		t.Fatal("expected SQS message ID")
	}

	if receiptHandle == "" {
		t.Fatal("expected SQS receipt handle")
	}

	if message.Data.IdempotencyKey != "consumer-test-001" {
		t.Fatalf(
			"unexpected idempotency key: %s",
			message.Data.IdempotencyKey,
		)
	}

	if message.Data.Type != "BET" {
		t.Fatalf(
			"unexpected transaction type: %s",
			message.Data.Type,
		)
	}

	if message.Data.Amount != 1000 {
		t.Fatalf(
			"unexpected amount: %d",
			message.Data.Amount,
		)
	}

	if err := client.DeleteMessage(ctx, queueURL, receiptHandle); err != nil {
		t.Fatalf("failed to delete message: %v", err)
	}

	if body == "" {
		t.Fatal("expected raw message body")
	}
}

func TestBuildWagerTransactionFromMessage(t *testing.T) {
	message := WagerTransactionMessage{
		MessageID:     "message-1",
		CorrelationID: "correlation-1",
		Data: WagerTransactionMessageData{
			IdempotencyKey:        "idem-1",
			TransactionID:         "transaction-1",
			PlayerID:              "player-1",
			ProviderID:            "provider-1",
			ExternalTransactionID: "external-1",
			WalletID:              "wallet-1",
			RoundID:               "round-1",
			GameID:                "game-1",
			Type:                  "BET",
			Amount:                100,
			Currency:              "BRL",
		},
	}

	now := time.Now()
	payloadHash := "test-payload-hash"

	transaction, err := buildWagerTransactionFromMessage(
		message,
		payloadHash,
		now,
	)
	if err != nil {
		t.Fatal(err)
	}

	if transaction.ID != "transaction-1" {
		t.Fatalf("expected transaction ID transaction-1, got %s", transaction.ID)
	}

	if transaction.IdempotencyKey != "idem-1" {
		t.Fatalf("expected idempotency key idem-1, got %s", transaction.IdempotencyKey)
	}

	if transaction.PayloadHash != payloadHash {
		t.Fatalf("expected payload hash %s, got %s", payloadHash, transaction.PayloadHash)
	}

	if transaction.Type != domain.TransactionBet {
		t.Fatalf("expected transaction type BET, got %s", transaction.Type)
	}

	if transaction.Amount.Amount != 100 {
		t.Fatalf("expected amount 100, got %d", transaction.Amount.Amount)
	}

	if transaction.Amount.Currency != "BRL" {
		t.Fatalf("expected currency BRL, got %s", transaction.Amount.Currency)
	}

	if transaction.State != domain.TransactionPending {
		t.Fatalf("expected transaction state PENDING, got %s", transaction.State)
	}
}

func TestHashPayload(t *testing.T) {
	body := `{"transactionId":"transaction-1","amount":100,"currency":"BRL"}`

	hash := hashPayload(body)

	if hash == "" {
		t.Fatal("expected payload hash")
	}

	if len(hash) != 64 {
		t.Fatalf("expected SHA-256 hash with 64 characters, got %d", len(hash))
	}

	if hash != hashPayload(body) {
		t.Fatal("expected same payload to produce the same hash")
	}

	otherHash := hashPayload(`{"transactionId":"transaction-2","amount":100,"currency":"BRL"}`)

	if hash == otherHash {
		t.Fatal("expected different payloads to produce different hashes")
	}
}
