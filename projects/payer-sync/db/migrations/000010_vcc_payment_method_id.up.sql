-- +goose Up

ALTER TABLE vcc_payment_groups
    ADD COLUMN IF NOT EXISTS payment_method_id TEXT;

-- +goose Down

ALTER TABLE vcc_payment_groups
    DROP COLUMN IF EXISTS payment_method_id;
