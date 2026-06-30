-- +goose Up

CREATE TABLE IF NOT EXISTS reconciled_payments (
    reconciled_payment_id TEXT PRIMARY KEY,
    location_id TEXT NOT NULL,
    era_payment_group_id TEXT NOT NULL UNIQUE REFERENCES era_payment_groups(group_id),
    vcc_payment_group_id TEXT NOT NULL UNIQUE REFERENCES vcc_payment_groups(group_id),
    trace_number TEXT NOT NULL,
    matched_amount NUMERIC(12, 2) NOT NULL,
    payer_name TEXT,
    provider_npi TEXT,
    provider_tax_id TEXT,
    status TEXT NOT NULL CHECK (
        status IN (
            'MATCHED',
            'PROCESSING_PAYMENT',
            'PROCESSING_FAILED',
            'PAYMENT_SUCCEEDED',
            'PAYMENT_FAILED',
            'WRITING_BACK',
            'POSTED',
            'PARTIALLY_POSTED',
            'WRITEBACK_FAILED',
            'NOTIFIED',
            'EXCEPTION'
        )
    ),
    matched_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_reconciled_payments_location_trace
    ON reconciled_payments (location_id, trace_number);

CREATE INDEX IF NOT EXISTS idx_reconciled_payments_status_matched_at
    ON reconciled_payments (status, matched_at DESC);

ALTER TABLE era_payment_groups
    ADD COLUMN IF NOT EXISTS first_received_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS matched_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS exception_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS prior_status TEXT,
    ADD COLUMN IF NOT EXISTS exception_reason TEXT;

UPDATE era_payment_groups
SET first_received_at = created_at
WHERE first_received_at IS NULL;

ALTER TABLE era_payment_groups
    ALTER COLUMN first_received_at SET DEFAULT NOW(),
    ALTER COLUMN first_received_at SET NOT NULL;

ALTER TABLE era_payment_groups
    DROP CONSTRAINT IF EXISTS era_payment_groups_status_check;

ALTER TABLE era_payment_groups
    ADD CONSTRAINT era_payment_groups_status_check
    CHECK (status IN ('AWAITING_VCC', 'MATCHED', 'EXCEPTION', 'EXCEPTION_UNMATCHED'));

DROP INDEX IF EXISTS idx_era_pg_location_trace_active;

CREATE UNIQUE INDEX IF NOT EXISTS idx_era_pg_location_trace_active
    ON era_payment_groups (location_id, trace_number)
    WHERE status NOT IN ('EXCEPTION', 'EXCEPTION_UNMATCHED');

CREATE INDEX IF NOT EXISTS idx_era_pg_status_first_received
    ON era_payment_groups (status, first_received_at);

ALTER TABLE vcc_payment_groups
    ADD COLUMN IF NOT EXISTS first_received_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS matched_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS exception_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS prior_status TEXT,
    ADD COLUMN IF NOT EXISTS exception_reason TEXT;

UPDATE vcc_payment_groups
SET first_received_at = created_at
WHERE first_received_at IS NULL;

ALTER TABLE vcc_payment_groups
    ALTER COLUMN first_received_at SET DEFAULT NOW(),
    ALTER COLUMN first_received_at SET NOT NULL;

ALTER TABLE vcc_payment_groups
    DROP CONSTRAINT IF EXISTS vcc_payment_groups_status_check;

ALTER TABLE vcc_payment_groups
    ADD CONSTRAINT vcc_payment_groups_status_check
    CHECK (status IN ('AWAITING_ERA', 'MATCHED', 'EXCEPTION', 'EXCEPTION_UNMATCHED'));

CREATE INDEX IF NOT EXISTS idx_vcc_pg_status_first_received
    ON vcc_payment_groups (status, first_received_at);
