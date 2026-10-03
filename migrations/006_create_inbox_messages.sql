CREATE TABLE inbox_messages (
    message_id TEXT PRIMARY KEY,
    idempotency_key TEXT NOT NULL,
    received_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    processed_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX idx_inbox_messages_idempotency_key
ON inbox_messages (idempotency_key);