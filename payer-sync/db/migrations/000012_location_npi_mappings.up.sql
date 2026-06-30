-- +goose Up
CREATE TABLE IF NOT EXISTS location_npi_mappings (
    location_id TEXT PRIMARY KEY,
    provider_npi TEXT NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (LENGTH(BTRIM(location_id)) > 0),
    CHECK (LENGTH(BTRIM(provider_npi)) > 0)
);

CREATE INDEX IF NOT EXISTS idx_location_npi_mappings_enabled
    ON location_npi_mappings (enabled)
    WHERE enabled = TRUE;

-- +goose Down
DROP TABLE IF EXISTS location_npi_mappings;
