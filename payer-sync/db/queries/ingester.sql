-- name: CreateERARemittance :one
INSERT INTO era_remittances (
    era_id,
    location_id,
    payer_name,
    provider_npi,
    provider_tax_id,
    bpr_amount,
    payment_method,
    trace_number,
    status,
    received_at,
    file_hash,
    raw_storage_key
) VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7,
    $8,
    $9,
    $10,
    $11,
    $12
)
RETURNING *;

-- name: UpdateERARemittanceParsed :exec
UPDATE era_remittances
SET payer_name = $2,
    provider_npi = $3,
    provider_tax_id = $4,
    bpr_amount = $5,
    payment_method = $6,
    trace_number = $7,
    status = 'PARSED',
    updated_at = NOW()
WHERE era_id = $1;

-- name: SetERARemittanceParseFailure :exec
UPDATE era_remittances
SET status = 'EXCEPTION_PARSE_FAILED',
    updated_at = NOW()
WHERE era_id = $1;

-- name: GetERARemittance :one
SELECT *
FROM era_remittances
WHERE era_id = $1
LIMIT 1;

-- name: GetERARemittanceByLocationAndHash :one
SELECT *
FROM era_remittances
WHERE location_id = $1
  AND file_hash = $2
LIMIT 1;

-- name: ListERARemittancesByReceivedAt :many
SELECT *
FROM era_remittances
ORDER BY received_at DESC
LIMIT $1;

-- name: CreateVCCFile :one
INSERT INTO vcc_files (
    vcc_file_id,
    location_id,
    received_at,
    file_hash,
    raw_storage_key,
    row_count,
    source_filename,
    status
) VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7,
    $8
)
RETURNING *;

-- name: UpdateVCCFileParsed :exec
UPDATE vcc_files
SET row_count = $2,
    status = 'PARSED',
    updated_at = NOW()
WHERE vcc_file_id = $1;

-- name: SetVCCFileParseFailure :exec
UPDATE vcc_files
SET row_count = $2,
    status = 'EXCEPTION_PARSE_FAILED',
    updated_at = NOW()
WHERE vcc_file_id = $1;

-- name: GetVCCFile :one
SELECT *
FROM vcc_files
WHERE vcc_file_id = $1
LIMIT 1;

-- name: GetVCCFileByLocationAndHash :one
SELECT *
FROM vcc_files
WHERE location_id = $1
  AND file_hash = $2
LIMIT 1;

-- name: CreateJobRun :one
INSERT INTO job_runs (
    run_id,
    job_type,
    started_at,
    finished_at,
    status,
    files_processed,
    records_matched,
    errors
) VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7,
    $8
)
RETURNING *;

-- name: UpdateJobRunResult :exec
UPDATE job_runs
SET finished_at = $2,
    status = $3,
    files_processed = $4,
    records_matched = $5,
    errors = $6
WHERE run_id = $1;

-- name: ListEnabledIngestionTargets :many
SELECT location_id, provider_npi
FROM location_npi_mappings
WHERE enabled = TRUE
ORDER BY location_id;

-- name: ListEnabledIngestionTargetsForLocation :many
SELECT location_id, provider_npi
FROM location_npi_mappings
WHERE enabled = TRUE
    AND location_id = $1
ORDER BY location_id;

-- name: GetActiveERAPaymentGroup :one
SELECT group_id, era_id, location_id, trace_number, bpr_amount::text AS bpr_amount, claim_count, claims,
       adjustments, status,
       reconciliation_triggered_at, created_at
FROM era_payment_groups
WHERE location_id = $1 AND trace_number = $2 AND status NOT IN ('EXCEPTION', 'EXCEPTION_UNMATCHED')
LIMIT 1;

-- name: GetMatchingERAForVCC :one
SELECT group_id, era_id, location_id, trace_number, bpr_amount::text AS bpr_amount, claim_count, claims,
       adjustments, status,
       reconciliation_triggered_at, created_at
FROM era_payment_groups
WHERE location_id = $1 AND trace_number = $2 AND status = 'AWAITING_VCC'
LIMIT 1;

-- name: CreateERAPaymentGroup :exec
INSERT INTO era_payment_groups (
    group_id,
    era_id,
    location_id,
    trace_number,
    bpr_amount,
    claim_count,
    claims,
    adjustments,
    status
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9);

-- name: SetERAPaymentGroupException :exec
UPDATE era_payment_groups SET status = 'EXCEPTION', updated_at = NOW() WHERE group_id = $1;

-- name: MarkERAReconciliationTriggered :exec
UPDATE era_payment_groups SET reconciliation_triggered_at = NOW(), updated_at = NOW() WHERE group_id = $1;

-- name: GetActiveVCCPaymentGroup :one
SELECT group_id, vcc_file_id, location_id, trace_id, payment_id, provider_npi,
       provider_tax_id, card_fingerprint, total_amount::text AS total_amount, status,
       is_authoritative, reconciliation_triggered_at, created_at
FROM vcc_payment_groups
WHERE location_id = $1 AND trace_id = $2 AND card_fingerprint = $3 AND is_authoritative = true AND status NOT IN ('EXCEPTION', 'EXCEPTION_UNMATCHED')
LIMIT 1;

-- name: GetMatchingVCCForERA :one
SELECT group_id, vcc_file_id, location_id, trace_id, payment_id, provider_npi,
       provider_tax_id, card_fingerprint, total_amount::text AS total_amount, status,
       is_authoritative, reconciliation_triggered_at, created_at
FROM vcc_payment_groups
WHERE location_id = $1 AND trace_id = $2 AND status = 'AWAITING_ERA' AND is_authoritative = true
LIMIT 1;

-- name: CreateVCCPaymentGroup :exec
INSERT INTO vcc_payment_groups
    (group_id, vcc_file_id, location_id, trace_id, payment_id, provider_npi, provider_tax_id,
     card_fingerprint, total_amount, status, is_authoritative, payment_method_id)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12);

-- name: SetVCCPaymentGroupException :exec
UPDATE vcc_payment_groups SET status = 'EXCEPTION', updated_at = NOW() WHERE group_id = $1;

-- name: MarkVCCGroupNonAuthoritative :exec
UPDATE vcc_payment_groups SET is_authoritative = false, updated_at = NOW() WHERE group_id = $1;

-- name: MarkVCCReconciliationTriggered :exec
UPDATE vcc_payment_groups SET reconciliation_triggered_at = NOW(), updated_at = NOW() WHERE group_id = $1;

-- name: InsertVCCRow :exec
INSERT INTO vcc_rows
    (row_id, vcc_file_id, vcc_payment_group_id, location_id, payment_id, trace_id,
     payer_name, provider_npi, provider_tax_id, issue_date, amount,
     card_fingerprint, last4, expiration_date, patient_id, claim_id,
     service_date_start, service_date_end)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)
ON CONFLICT DO NOTHING;

-- name: InsertStateTransition :one
INSERT INTO state_transitions (
    transition_id,
    entity_type,
    entity_id,
    from_state,
    to_state,
    transitioned_at,
    reason
) VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7
)
RETURNING *;
