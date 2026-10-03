CREATE TABLE outbox_events (
    id TEXT PRIMARY KEY,
    event_type TEXT NOT NULL,
    aggregate_id TEXT NOT NULL,
    correlation_id TEXT,
    causation_id TEXT,
    occurred_at TIMESTAMPTZ NOT NULL,
    version INTEGER NOT NULL,
    payload JSONB NOT NULL,
    attempts INTEGER NOT NULL DEFAULT 0,
    next_attempt_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    published_at TIMESTAMPTZ
);

CREATE INDEX idx_outbox_events_pending
ON outbox_events (next_attempt_at)
WHERE published_at IS NULL;