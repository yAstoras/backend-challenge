package application

import (
	"context"
	"encoding/json"
	"testing"
)

func TestSQSClientSendReceiveAndDeleteMessage(t *testing.T) {
	ctx := context.Background()

	client, err := NewSQSClient()
	if err != nil {
		t.Fatalf("failed to create SQS client: %v", err)
	}

	queueURL, err := client.QueueURL(ctx, "wager-transactions")
	if err != nil {
		t.Fatalf("failed to get queue URL: %v", err)
	}

	body := `{"test":"hello-sqs"}`

	err = client.SendMessage(ctx, queueURL, body)
	if err != nil {
		t.Fatalf("failed to send message: %v", err)
	}

	messageID, receivedBody, receiptHandle, err := client.ReceiveMessage(ctx, queueURL)
	if err != nil {
		t.Fatalf("failed to receive message: %v", err)
	}

	if messageID == "" {
		t.Fatal("message ID is empty")
	}

	if receivedBody != body {
		t.Fatalf(
			"unexpected message body: got %q, want %q",
			receivedBody,
			body,
		)
	}

	if receiptHandle == "" {
		t.Fatal("receipt handle is empty")
	}

	err = client.DeleteMessage(ctx, queueURL, receiptHandle)
	if err != nil {
		t.Fatalf("failed to delete message: %v", err)
	}

	t.Logf("message ID: %s", messageID)
	t.Log("message sent, received and deleted successfully")
}

func TestWagerTransactionMessageJSON(t *testing.T) {
	body := `{
		"messageId": "message-123",
		"correlationId": "correlation-123",
		"data": {
			"idempotencyKey": "idempotency-123",
			"transactionId": "transaction-123",
			"playerId": "player-123",
			"providerId": "provider-123",
			"externalTransactionId": "external-123",
			"walletId": "wallet-123",
			"roundId": "round-123",
			"gameId": "game-123",
			"type": "BET",
			"amount": 1000,
			"currency": "BRL"
		}
	}`

	var message WagerTransactionMessage

	err := json.Unmarshal([]byte(body), &message)
	if err != nil {
		t.Fatalf("failed to unmarshal message: %v", err)
	}

	if message.MessageID != "message-123" {
		t.Fatalf("unexpected message ID: %s", message.MessageID)
	}

	if message.Data.IdempotencyKey != "idempotency-123" {
		t.Fatalf(
			"unexpected idempotency key: %s",
			message.Data.IdempotencyKey,
		)
	}

	if message.Data.Type != "BET" {
		t.Fatalf("unexpected transaction type: %s", message.Data.Type)
	}

	if message.Data.Amount != 1000 {
		t.Fatalf("unexpected amount: %d", message.Data.Amount)
	}

	if message.Data.Currency != "BRL" {
		t.Fatalf("unexpected currency: %s", message.Data.Currency)
	}

	t.Log("SQS message JSON parsed successfully")
}
