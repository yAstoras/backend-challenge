CREATE TABLE wager_transactions (
    id TEXT PRIMARY KEY,

    provider_id TEXT NOT NULL,
    external_transaction_id TEXT NOT NULL,

    idempotency_key TEXT NOT NULL,
    payload_hash TEXT NOT NULL,

    player_id TEXT NOT NULL,
    wallet_id TEXT NOT NULL,
    round_id TEXT NOT NULL,
    game_id TEXT NOT NULL,

    transaction_type VARCHAR(20) NOT NULL,
    amount BIGINT NOT NULL,
    currency VARCHAR(3) NOT NULL,

    reference_external_transaction_id TEXT,

    state VARCHAR(30) NOT NULL,

    failure_code TEXT,

    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,

    CONSTRAINT wager_transactions_amount_valid
        CHECK (amount > 0 OR transaction_type = 'LOSS'),

    CONSTRAINT wager_transactions_type_valid
        CHECK (
            transaction_type IN (
                'BET',
                'WIN',
                'LOSS',
                'REFUND',
                'ROLLBACK'
            )
        ),

    CONSTRAINT wager_transactions_state_valid
        CHECK (
            state IN (
                'PENDING',
                'PENDING_REFERENCE',
                'PROCESSED',
                'REJECTED',
                'FAILED'
            )
        ),

    CONSTRAINT wager_transactions_provider_external_unique
        UNIQUE (provider_id, external_transaction_id),

    CONSTRAINT wager_transactions_idempotency_unique
        UNIQUE (idempotency_key)
);