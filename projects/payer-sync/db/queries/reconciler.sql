-- name: ListUnmatchedERAPaymentGroups :many
SELECT
    epg.group_id,
    epg.era_id,
    epg.location_id,
    epg.trace_number,
    epg.bpr_amount::text AS bpr_amount,
    epg.status,
    epg.reconciliation_triggered_at,
    epg.created_at,
    epg.updated_at,
    epg.first_received_at,
    epg.matched_at,
    epg.exception_at,
    epg.prior_status,
    epg.exception_reason,
    er.payer_name,
    er.provider_npi,
    er.provider_tax_id
FROM era_payment_groups epg
JOIN era_remittances er ON er.era_id = epg.era_id
WHERE epg.status = 'AWAITING_VCC'
ORDER BY epg.first_received_at ASC, epg.group_id ASC;

-- name: ListUnmatchedVCCPaymentGroups :many
SELECT
    group_id,
    vcc_file_id,
    location_id,
    trace_id,
    payment_id,
    provider_npi,
    provider_tax_id,
    card_fingerprint,
    total_amount::text AS total_amount,
    status,
    is_authoritative,
    reconciliation_triggered_at,
    created_at,
    updated_at,
    first_received_at,
    matched_at,
    exception_at,
    prior_status,
    exception_reason
FROM vcc_payment_groups
WHERE status = 'AWAITING_ERA'
  AND is_authoritative = TRUE
ORDER BY first_received_at ASC, group_id ASC;

-- name: ListVCCCounterpartCandidates :many
SELECT
    group_id,
    vcc_file_id,
    location_id,
    trace_id,
    payment_id,
    provider_npi,
    provider_tax_id,
    card_fingerprint,
    total_amount::text AS total_amount,
    status,
    is_authoritative,
    reconciliation_triggered_at,
    created_at,
    updated_at,
    first_received_at,
    matched_at,
    exception_at,
    prior_status,
    exception_reason
FROM vcc_payment_groups
WHERE location_id = $1
  AND trace_id = $2
  AND status = 'AWAITING_ERA'
  AND is_authoritative = TRUE
ORDER BY first_received_at ASC, group_id ASC;

-- name: ListERACounterpartCandidates :many
SELECT
    epg.group_id,
    epg.era_id,
    epg.location_id,
    epg.trace_number,
    epg.bpr_amount::text AS bpr_amount,
    epg.status,
    epg.reconciliation_triggered_at,
    epg.created_at,
    epg.updated_at,
    epg.first_received_at,
    epg.matched_at,
    epg.exception_at,
    epg.prior_status,
    epg.exception_reason,
    er.payer_name,
    er.provider_npi,
    er.provider_tax_id
FROM era_payment_groups epg
JOIN era_remittances er ON er.era_id = epg.era_id
WHERE epg.location_id = $1
  AND epg.trace_number = $2
  AND epg.status = 'AWAITING_VCC'
ORDER BY epg.first_received_at ASC, epg.group_id ASC;

-- name: CreateReconciledPayment :one
INSERT INTO reconciled_payments (
    reconciled_payment_id,
    location_id,
    era_payment_group_id,
    vcc_payment_group_id,
    trace_number,
    matched_amount,
    payer_name,
    provider_npi,
    provider_tax_id,
    status,
    matched_at
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
    $11
)
RETURNING
    reconciled_payment_id,
    location_id,
    era_payment_group_id,
    vcc_payment_group_id,
    trace_number,
    matched_amount,
    payer_name,
    provider_npi,
    provider_tax_id,
    status,
    matched_at,
    created_at,
    updated_at;

-- name: MarkERAPaymentGroupMatched :one
UPDATE era_payment_groups
SET status = 'MATCHED',
    matched_at = $2,
    updated_at = $2
WHERE group_id = $1
  AND status = 'AWAITING_VCC'
RETURNING group_id, status, matched_at;

-- name: MarkVCCPaymentGroupMatched :one
UPDATE vcc_payment_groups
SET status = 'MATCHED',
    matched_at = $2,
    updated_at = $2
WHERE group_id = $1
  AND status = 'AWAITING_ERA'
RETURNING group_id, status, matched_at;

-- name: NotifyReconciledPaymentMatched :exec
SELECT pg_notify('reconciled_payment_matched', $1);

-- name: ExpireUnmatchedERAPaymentGroup :one
UPDATE era_payment_groups
SET prior_status = status,
    status = 'EXCEPTION_UNMATCHED',
    exception_at = $2,
    exception_reason = $3,
    updated_at = $2
WHERE group_id = $1
  AND status = 'AWAITING_VCC'
RETURNING group_id, prior_status, status, exception_at, exception_reason;

-- name: ExpireUnmatchedVCCPaymentGroup :one
UPDATE vcc_payment_groups
SET prior_status = status,
    status = 'EXCEPTION_UNMATCHED',
    exception_at = $2,
    exception_reason = $3,
    updated_at = $2
WHERE group_id = $1
  AND status = 'AWAITING_ERA'
RETURNING group_id, prior_status, status, exception_at, exception_reason;

-- name: InsertReconcilerStateTransition :one
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

-- name: CreateReconcilerJobRun :one
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
    'reconciler',
    $2,
    $3,
    $4,
    $5,
    $6,
    $7
)
RETURNING *;

-- name: UpdateReconcilerJobRunResult :exec
UPDATE job_runs
SET finished_at = $2,
    status = $3,
    files_processed = $4,
    records_matched = $5,
    errors = $6
WHERE run_id = $1
  AND job_type = 'reconciler';
