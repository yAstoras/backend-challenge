package application

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OutboxEvent struct {
	ID            string
	EventType     string
	AggregateID   string
	CorrelationID string
	CausationID   string
	OccurredAt    time.Time
	Version       int
	Payload       any
	Attempts      int
	NextAttemptAt time.Time
	PublishedAt   *time.Time
}

type OutboxRepository struct {
	db  *pgxpool.Pool
	sqs *SQSClient
}

func NewOutboxRepository(
	db *pgxpool.Pool,
	sqs *SQSClient,
) *OutboxRepository {
	return &OutboxRepository{
		db:  db,
		sqs: sqs,
	}
}

func (r *OutboxRepository) CreateTx(
	ctx context.Context,
	tx pgx.Tx,
	event OutboxEvent,
) error {
	payload, err := json.Marshal(event.Payload)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO outbox_events (
			id,
			event_type,
			aggregate_id,
			correlation_id,
			causation_id,
			occurred_at,
			version,
			payload,
			attempts,
			next_attempt_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (id) DO NOTHING
	`,
		event.ID,
		event.EventType,
		event.AggregateID,
		event.CorrelationID,
		event.CausationID,
		event.OccurredAt,
		event.Version,
		payload,
		event.Attempts,
		event.NextAttemptAt,
	)

	return err
}

func (r *OutboxRepository) FindPending(
	ctx context.Context,
	limit int,
) ([]OutboxEvent, error) {
	rows, err := r.db.Query(ctx, `
		SELECT
			id,
			event_type,
			aggregate_id,
			COALESCE(correlation_id, ''),
			COALESCE(causation_id, ''),
			occurred_at,
			version,
			payload,
			attempts,
			next_attempt_at,
			published_at
		FROM outbox_events
		WHERE published_at IS NULL
		  AND next_attempt_at <= NOW()
		ORDER BY occurred_at ASC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []OutboxEvent

	for rows.Next() {
		var event OutboxEvent
		var payload []byte

		if err := rows.Scan(
			&event.ID,
			&event.EventType,
			&event.AggregateID,
			&event.CorrelationID,
			&event.CausationID,
			&event.OccurredAt,
			&event.Version,
			&payload,
			&event.Attempts,
			&event.NextAttemptAt,
			&event.PublishedAt,
		); err != nil {
			return nil, err
		}

		event.Payload = payload
		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return events, nil
}

func (r *OutboxRepository) MarkPublished(
	ctx context.Context,
	eventID string,
) error {
	_, err := r.db.Exec(ctx, `
		UPDATE outbox_events
		SET published_at = NOW()
		WHERE id = $1
		  AND published_at IS NULL
	`, eventID)

	return err
}

func (r *OutboxRepository) PublishPending(
	ctx context.Context,
	limit int,
) error {
	events, err := r.FindPending(ctx, limit)
	if err != nil {
		return err
	}

	for _, event := range events {
		if err := r.publishEvent(ctx, event); err != nil {
			if markErr := r.MarkFailed(ctx, event.ID); markErr != nil {
				return markErr
			}

			continue
		}

		if err := r.MarkPublished(ctx, event.ID); err != nil {
			return err
		}
	}

	return nil
}

func (r *OutboxRepository) publishEvent(
	ctx context.Context,
	event OutboxEvent,
) error {
	if r.sqs == nil {
		return errors.New("outbox SQS client is nil")
	}

	queueURL, err := r.sqs.QueueURL(ctx, "wager-events")
	if err != nil {
		return err
	}

	var payload json.RawMessage

	switch value := event.Payload.(type) {
	case []byte:
		payload = json.RawMessage(value)

	case json.RawMessage:
		payload = value

	default:
		data, err := json.Marshal(value)
		if err != nil {
			return err
		}

		payload = json.RawMessage(data)
	}

	envelope := map[string]any{
		"eventId":       event.ID,
		"eventType":     event.EventType,
		"aggregateId":   event.AggregateID,
		"correlationId": event.CorrelationID,
		"occurredAt":    event.OccurredAt,
		"version":       event.Version,
		"data":          json.RawMessage(payload),
	}

	if event.CausationID != "" {
		envelope["causationId"] = event.CausationID
	}

	body, err := json.Marshal(envelope)
	if err != nil {
		return err
	}

	return r.sqs.SendMessage(ctx, queueURL, string(body))
}

func nextOutboxRetryAt(attempts int, now time.Time) time.Time {
	delay := time.Second

	for i := 1; i < attempts; i++ {
		delay *= 2
	}

	return now.Add(delay)
}

func (r *OutboxRepository) MarkFailed(
	ctx context.Context,
	eventID string,
) error {
	var attempts int

	err := r.db.QueryRow(ctx, `
		SELECT attempts
		FROM outbox_events
		WHERE id = $1
		  AND published_at IS NULL
	`, eventID).Scan(&attempts)

	if err != nil {
		return err
	}

	nextAttempt := nextOutboxRetryAt(attempts+1, time.Now())

	_, err = r.db.Exec(ctx, `
		UPDATE outbox_events
		SET
			attempts = attempts + 1,
			next_attempt_at = $2
		WHERE id = $1
		  AND published_at IS NULL
	`,
		eventID,
		nextAttempt,
	)

	return err
}
