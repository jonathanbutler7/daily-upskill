-- +goose Up
CREATE TABLE IF NOT EXISTS job_runs (
    run_id TEXT PRIMARY KEY,
    job_type TEXT NOT NULL CHECK (job_type IN ('ingester', 'reconciler', 'processor', 'writeback')),
    started_at TIMESTAMPTZ NOT NULL,
    finished_at TIMESTAMPTZ,
    status TEXT NOT NULL CHECK (status IN ('success', 'failure', 'partial')),
    files_processed INTEGER NOT NULL DEFAULT 0 CHECK (files_processed >= 0),
    records_matched INTEGER NOT NULL DEFAULT 0 CHECK (records_matched >= 0),
    errors JSONB NOT NULL DEFAULT '[]'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_job_runs_job_type_started_at
    ON job_runs (job_type, started_at DESC);

CREATE INDEX IF NOT EXISTS idx_job_runs_status_started_at
    ON job_runs (status, started_at DESC);

CREATE TABLE IF NOT EXISTS state_transitions (
    transition_id TEXT PRIMARY KEY,
    entity_type TEXT NOT NULL CHECK (
        entity_type IN (
            'era',
            'vcc',
            'era_remittance',
            'vcc_file',
            'era_payment_group',
            'vcc_payment_group',
            'reconciled_payment',
            'ledger_posting'
        )
    ),
    entity_id TEXT NOT NULL,
    from_state TEXT,
    to_state TEXT NOT NULL,
    transitioned_at TIMESTAMPTZ NOT NULL,
    reason TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_state_transitions_entity
    ON state_transitions (entity_type, entity_id, transitioned_at DESC);

CREATE INDEX IF NOT EXISTS idx_state_transitions_to_state
    ON state_transitions (to_state, transitioned_at DESC);

-- +goose Down
DROP TABLE IF EXISTS state_transitions;
DROP TABLE IF EXISTS job_runs;
