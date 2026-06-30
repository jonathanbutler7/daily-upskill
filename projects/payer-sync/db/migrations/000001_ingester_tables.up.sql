-- +goose Up
CREATE TABLE IF NOT EXISTS era_remittances (
    era_id TEXT PRIMARY KEY,
    location_id TEXT NOT NULL,
    payer_name TEXT,
    provider_npi TEXT,
    provider_tax_id TEXT,
    bpr_amount NUMERIC(12, 2),
    payment_method TEXT,
    trace_number TEXT,
    status TEXT NOT NULL CHECK (status IN ('RECEIVED_RAW', 'PARSED', 'EXCEPTION_PARSE_FAILED')),
    received_at TIMESTAMPTZ NOT NULL,
    file_hash TEXT NOT NULL,
    raw_storage_key TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (location_id, file_hash)
);

CREATE INDEX IF NOT EXISTS idx_era_remittances_trace_number
    ON era_remittances (trace_number);

CREATE INDEX IF NOT EXISTS idx_era_remittances_received_at
    ON era_remittances (received_at DESC);

CREATE TABLE IF NOT EXISTS vcc_files (
    vcc_file_id TEXT PRIMARY KEY,
    location_id TEXT NOT NULL,
    received_at TIMESTAMPTZ NOT NULL,
    file_hash TEXT NOT NULL,
    raw_storage_key TEXT NOT NULL,
    row_count INTEGER NOT NULL CHECK (row_count >= 0),
    source_filename TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (location_id, file_hash)
);

CREATE INDEX IF NOT EXISTS idx_vcc_files_received_at
    ON vcc_files (received_at DESC);

-- +goose Down
DROP TABLE IF EXISTS vcc_files;
DROP TABLE IF EXISTS era_remittances;
