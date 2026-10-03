CREATE TABLE wallet_ledger (
    id TEXT PRIMARY KEY,
    wallet_id TEXT NOT NULL,
    transaction_id TEXT NOT NULL,
    direction VARCHAR(10) NOT NULL,
    amount BIGINT NOT NULL,
    balance_before BIGINT NOT NULL,
    balance_after BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,

    CONSTRAINT wallet_ledger_amount_positive
        CHECK (amount > 0),

    CONSTRAINT wallet_ledger_balance_before_non_negative
        CHECK (balance_before >= 0),

    CONSTRAINT wallet_ledger_balance_after_non_negative
        CHECK (balance_after >= 0),

    CONSTRAINT wallet_ledger_direction_valid
        CHECK (direction IN ('DEBIT', 'CREDIT')),

    CONSTRAINT wallet_ledger_transaction_unique
        UNIQUE (wallet_id, transaction_id)
);