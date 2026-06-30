-- +goose Up

ALTER TABLE reconciled_payments
    ADD COLUMN IF NOT EXISTS idempotency_key TEXT UNIQUE,
    ADD COLUMN IF NOT EXISTS processor_payment_intent_id TEXT,
    ADD COLUMN IF NOT EXISTS processor_charge_id TEXT,
    ADD COLUMN IF NOT EXISTS processor_error_code TEXT,
    ADD COLUMN IF NOT EXISTS processor_error_message TEXT,
    ADD COLUMN IF NOT EXISTS processing_started_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS processing_completed_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS retry_count INT NOT NULL DEFAULT 0;

ALTER TABLE vcc_payment_groups
    ADD COLUMN IF NOT EXISTS version INT NOT NULL DEFAULT 1;

CREATE TABLE IF NOT EXISTS processor_attempts (
    attempt_id              TEXT PRIMARY KEY,
    reconciled_payment_id   TEXT NOT NULL REFERENCES reconciled_payments(reconciled_payment_id),
    idempotency_key         TEXT NOT NULL,
    attempt_number          INT NOT NULL,
    outcome                 TEXT NOT NULL CHECK (outcome IN ('succeeded', 'failed', 'retrying')),
    error_code              TEXT,
    error_message           TEXT,
    attempted_at            TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_processor_attempts_payment
    ON processor_attempts (reconciled_payment_id, attempted_at DESC);

-- +goose Down

DROP INDEX IF EXISTS idx_processor_attempts_payment;
DROP TABLE IF EXISTS processor_attempts;

ALTER TABLE vcc_payment_groups
    DROP COLUMN IF EXISTS version;

ALTER TABLE reconciled_payments
    DROP COLUMN IF EXISTS idempotency_key,
    DROP COLUMN IF EXISTS processor_payment_intent_id,
    DROP COLUMN IF EXISTS processor_charge_id,
    DROP COLUMN IF EXISTS processor_error_code,
    DROP COLUMN IF EXISTS processor_error_message,
    DROP COLUMN IF EXISTS processing_started_at,
    DROP COLUMN IF EXISTS processing_completed_at,
    DROP COLUMN IF EXISTS retry_count;
