ALTER TABLE wager_transactions
ADD COLUMN reference_attempts INTEGER NOT NULL DEFAULT 0,
ADD COLUMN next_reference_attempt_at TIMESTAMPTZ NOT NULL DEFAULT NOW();