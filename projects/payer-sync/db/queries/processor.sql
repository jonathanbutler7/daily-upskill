-- name: ListMatchedPaymentsForProcessing :many
SELECT * FROM reconciled_payments
WHERE status = 'MATCHED'
ORDER BY matched_at ASC
LIMIT $1;

-- name: GetMatchedPaymentWithVCCDetails :one
-- PCI BOUNDARY: This query returns the Stripe payment method ID tokenized at ingestion.
-- card_number and cvv are never stored or returned; tokenization happened in the ingester.
SELECT
    rp.*,
    vpg.version AS vcc_version,
    vpg.card_fingerprint,
    vpg.total_amount,
    vpg.payment_method_id
FROM reconciled_payments rp
JOIN vcc_payment_groups vpg ON vpg.group_id = rp.vcc_payment_group_id
WHERE rp.reconciled_payment_id = $1
LIMIT 1;

-- name: BeginProcessing :one
UPDATE reconciled_payments
SET status = 'PROCESSING_PAYMENT',
    idempotency_key = $2,
    processing_started_at = $3,
    updated_at = $3
WHERE reconciled_payment_id = $1
  AND status = 'MATCHED'
RETURNING *;

-- name: RequeueClaimedPayment :execrows
UPDATE reconciled_payments
SET status = 'MATCHED',
    processing_started_at = NULL,
    processing_completed_at = NULL,
    updated_at = $2
WHERE reconciled_payment_id = $1
  AND status = 'PROCESSING_PAYMENT';

-- name: MarkPaymentSucceeded :one
UPDATE reconciled_payments
SET status = 'PAYMENT_SUCCEEDED',
    processor_payment_intent_id = $2,
    processing_completed_at = $3,
    updated_at = $3
WHERE reconciled_payment_id = $1
  AND status = 'PROCESSING_PAYMENT'
RETURNING *;

-- name: MarkPaymentFailed :one
UPDATE reconciled_payments
SET status = 'PROCESSING_FAILED',
    processor_error_code = $2,
    processor_error_message = $3,
    retry_count = $4,
    processing_completed_at = $5,
    updated_at = $5
WHERE reconciled_payment_id = $1
RETURNING *;

-- name: InsertProcessorAttempt :one
INSERT INTO processor_attempts (
    attempt_id,
    reconciled_payment_id,
    idempotency_key,
    attempt_number,
    outcome,
    error_code,
    error_message,
    attempted_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: InsertProcessorStateTransition :one
INSERT INTO state_transitions (
    transition_id,
    entity_type,
    entity_id,
    from_state,
    to_state,
    transitioned_at,
    reason
) VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;
