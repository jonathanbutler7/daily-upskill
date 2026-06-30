-- +goose Up
CREATE TABLE IF NOT EXISTS era_payment_groups (
    group_id TEXT PRIMARY KEY,
    era_id TEXT NOT NULL REFERENCES era_remittances(era_id),
    location_id TEXT NOT NULL,
    trace_number TEXT NOT NULL,
    bpr_amount NUMERIC(12, 2) NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('AWAITING_VCC', 'MATCHED', 'EXCEPTION')),
    reconciliation_triggered_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Only one non-exception group per (location, trace). Conflicts go to EXCEPTION.
CREATE UNIQUE INDEX IF NOT EXISTS idx_era_pg_location_trace_active
    ON era_payment_groups (location_id, trace_number)
    WHERE status != 'EXCEPTION';

CREATE INDEX IF NOT EXISTS idx_era_pg_location_trace
    ON era_payment_groups (location_id, trace_number);

CREATE TABLE IF NOT EXISTS vcc_payment_groups (
    group_id TEXT PRIMARY KEY,
    vcc_file_id TEXT NOT NULL REFERENCES vcc_files(vcc_file_id),
    location_id TEXT NOT NULL,
    trace_id TEXT NOT NULL,
    payment_id TEXT NOT NULL,
    provider_npi TEXT,
    provider_tax_id TEXT,
    card_fingerprint TEXT NOT NULL,
    total_amount NUMERIC(12, 2) NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('AWAITING_ERA', 'MATCHED', 'EXCEPTION')),
    is_authoritative BOOLEAN NOT NULL DEFAULT TRUE,
    reconciliation_triggered_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_vcc_pg_location_trace
    ON vcc_payment_groups (location_id, trace_id);

CREATE TABLE IF NOT EXISTS vcc_rows (
    row_id TEXT PRIMARY KEY,
    vcc_file_id TEXT NOT NULL REFERENCES vcc_files(vcc_file_id),
    vcc_payment_group_id TEXT REFERENCES vcc_payment_groups(group_id),
    location_id TEXT NOT NULL,
    payment_id TEXT NOT NULL,
    trace_id TEXT NOT NULL,
    payer_name TEXT,
    provider_npi TEXT,
    provider_tax_id TEXT,
    issue_date DATE NOT NULL,
    amount NUMERIC(12, 2) NOT NULL,
    card_fingerprint TEXT NOT NULL,
    last4 TEXT NOT NULL,
    expiration_date TEXT,
    patient_id TEXT,
    claim_id TEXT,
    service_date_start DATE,
    service_date_end DATE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    -- Natural key dedup (ING-008)
    UNIQUE (location_id, payment_id, trace_id, claim_id, amount, provider_npi, provider_tax_id, issue_date)
);

-- +goose Down
DROP TABLE IF EXISTS vcc_rows;
DROP TABLE IF EXISTS vcc_payment_groups;
DROP TABLE IF EXISTS era_payment_groups;
