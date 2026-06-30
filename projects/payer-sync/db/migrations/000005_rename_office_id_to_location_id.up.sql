-- +goose Up
-- +goose StatementBegin
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'public'
          AND table_name = 'era_remittances'
          AND column_name = 'office_id'
    ) AND NOT EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'public'
          AND table_name = 'era_remittances'
          AND column_name = 'location_id'
    ) THEN
        ALTER TABLE era_remittances RENAME COLUMN office_id TO location_id;
    END IF;
END $$;
-- +goose StatementEnd

-- +goose StatementBegin
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'public'
          AND table_name = 'vcc_files'
          AND column_name = 'office_id'
    ) AND NOT EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'public'
          AND table_name = 'vcc_files'
          AND column_name = 'location_id'
    ) THEN
        ALTER TABLE vcc_files RENAME COLUMN office_id TO location_id;
    END IF;
END $$;
-- +goose StatementEnd

-- +goose StatementBegin
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'public'
          AND table_name = 'era_payment_groups'
          AND column_name = 'office_id'
    ) AND NOT EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'public'
          AND table_name = 'era_payment_groups'
          AND column_name = 'location_id'
    ) THEN
        ALTER TABLE era_payment_groups RENAME COLUMN office_id TO location_id;
    END IF;
END $$;
-- +goose StatementEnd

-- +goose StatementBegin
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'public'
          AND table_name = 'vcc_payment_groups'
          AND column_name = 'office_id'
    ) AND NOT EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'public'
          AND table_name = 'vcc_payment_groups'
          AND column_name = 'location_id'
    ) THEN
        ALTER TABLE vcc_payment_groups RENAME COLUMN office_id TO location_id;
    END IF;
END $$;
-- +goose StatementEnd

-- +goose StatementBegin
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'public'
          AND table_name = 'vcc_rows'
          AND column_name = 'office_id'
    ) AND NOT EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'public'
          AND table_name = 'vcc_rows'
          AND column_name = 'location_id'
    ) THEN
        ALTER TABLE vcc_rows RENAME COLUMN office_id TO location_id;
    END IF;
END $$;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'public'
          AND table_name = 'vcc_rows'
          AND column_name = 'location_id'
    ) AND NOT EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'public'
          AND table_name = 'vcc_rows'
          AND column_name = 'office_id'
    ) THEN
        ALTER TABLE vcc_rows RENAME COLUMN location_id TO office_id;
    END IF;
END $$;
-- +goose StatementEnd

-- +goose StatementBegin
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'public'
          AND table_name = 'vcc_payment_groups'
          AND column_name = 'location_id'
    ) AND NOT EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'public'
          AND table_name = 'vcc_payment_groups'
          AND column_name = 'office_id'
    ) THEN
        ALTER TABLE vcc_payment_groups RENAME COLUMN location_id TO office_id;
    END IF;
END $$;
-- +goose StatementEnd

-- +goose StatementBegin
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'public'
          AND table_name = 'era_payment_groups'
          AND column_name = 'location_id'
    ) AND NOT EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'public'
          AND table_name = 'era_payment_groups'
          AND column_name = 'office_id'
    ) THEN
        ALTER TABLE era_payment_groups RENAME COLUMN location_id TO office_id;
    END IF;
END $$;
-- +goose StatementEnd

-- +goose StatementBegin
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'public'
          AND table_name = 'vcc_files'
          AND column_name = 'location_id'
    ) AND NOT EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'public'
          AND table_name = 'vcc_files'
          AND column_name = 'office_id'
    ) THEN
        ALTER TABLE vcc_files RENAME COLUMN location_id TO office_id;
    END IF;
END $$;
-- +goose StatementEnd

-- +goose StatementBegin
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'public'
          AND table_name = 'era_remittances'
          AND column_name = 'location_id'
    ) AND NOT EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'public'
          AND table_name = 'era_remittances'
          AND column_name = 'office_id'
    ) THEN
        ALTER TABLE era_remittances RENAME COLUMN location_id TO office_id;
    END IF;
END $$;
-- +goose StatementEnd
