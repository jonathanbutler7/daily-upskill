-- +goose Up
ALTER TABLE state_transitions
    DROP CONSTRAINT IF EXISTS state_transitions_entity_type_check;

ALTER TABLE state_transitions
    ADD CONSTRAINT state_transitions_entity_type_check
    CHECK (
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
    );

-- +goose Down
DELETE FROM state_transitions
WHERE entity_type NOT IN ('era', 'vcc', 'reconciled_payment', 'ledger_posting');

ALTER TABLE state_transitions
    DROP CONSTRAINT IF EXISTS state_transitions_entity_type_check;

ALTER TABLE state_transitions
    ADD CONSTRAINT state_transitions_entity_type_check
    CHECK (entity_type IN ('era', 'vcc', 'reconciled_payment', 'ledger_posting'));
