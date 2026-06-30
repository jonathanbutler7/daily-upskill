-- PayerSync reliability audit report
--
-- Run each query independently in psql or your SQL client.
-- The goal is to measure how well deduping and matching are behaving in practice.

-- 1) File-level dedupe summary.
-- Returns one row per source table with the total persisted files and the number
-- of distinct (location_id, file_hash) pairs. If files is larger than unique_files,
-- some ingest attempts collided on the file hash key.
SELECT
    'file-level dedupe summary' AS check_name,
    'files should equal unique_files for each source' AS what_to_validate,
    'vcc' AS source,
    COUNT(*) AS files,
    COUNT(DISTINCT (location_id, file_hash)) AS unique_files
FROM vcc_files
UNION ALL
SELECT
    'file-level dedupe summary' AS check_name,
    'files should equal unique_files for each source' AS what_to_validate,
    'era' AS source,
    COUNT(*) AS files,
    COUNT(DISTINCT (location_id, file_hash)) AS unique_files
FROM era_remittances;

-- 2) File-level duplicate key inspection.
-- Returns any raw file keys that appear more than once in persisted data.
-- On a healthy system this should normally return zero rows because the schema
-- enforces UNIQUE(location_id, file_hash) on both raw file tables.
SELECT
    'file-level duplicate key inspection' AS check_name,
    'expect zero rows' AS what_to_validate,
    'vcc' AS source,
    location_id,
    file_hash,
    COUNT(*) AS rows_with_same_key,
    MIN(received_at) AS first_received_at,
    MAX(received_at) AS last_received_at
FROM vcc_files
GROUP BY location_id, file_hash
HAVING COUNT(*) > 1
UNION ALL
SELECT
    'file-level duplicate key inspection' AS check_name,
    'expect zero rows' AS what_to_validate,
    'era' AS source,
    location_id,
    file_hash,
    COUNT(*) AS rows_with_same_key,
    MIN(received_at) AS first_received_at,
    MAX(received_at) AS last_received_at
FROM era_remittances
GROUP BY location_id, file_hash
HAVING COUNT(*) > 1
ORDER BY source, location_id, file_hash;

-- 3) VCC row dedupe effectiveness by file.
-- Returns one row per VCC file with the parsed row count, the number of rows that
-- actually survived into vcc_rows, and the difference between the two.
-- A positive dropped_or_deduped value means rows were eliminated by the natural-key
-- dedupe constraint or by conflict handling during ingest.
WITH stored_rows AS (
    SELECT
        vcc_file_id,
        COUNT(*) AS stored_row_count
    FROM vcc_rows
    GROUP BY vcc_file_id
)
SELECT
    'vcc row dedupe effectiveness' AS check_name,
    'dropped_or_deduped should be 0 for clean files; larger values mean row collisions or conflict handling' AS what_to_validate,
    f.vcc_file_id,
    f.location_id,
    f.source_filename,
    f.row_count AS parsed_rows,
    COALESCE(s.stored_row_count, 0) AS stored_rows,
    f.row_count - COALESCE(s.stored_row_count, 0) AS dropped_or_deduped
FROM vcc_files f
LEFT JOIN stored_rows s USING (vcc_file_id)
ORDER BY dropped_or_deduped DESC, f.vcc_file_id;

-- 4) Active trace collisions.
-- Returns any location/trace combinations that still have more than one active
-- payment group. For ERA, active means not EXCEPTION or EXCEPTION_UNMATCHED.
-- For VCC, active means not EXCEPTION or EXCEPTION_UNMATCHED.
-- These rows are the strongest signal that the matching inputs are ambiguous.
SELECT
    'active trace collisions' AS check_name,
    'expect zero rows; any row means more than one active group shares the same trace key' AS what_to_validate,
    'era' AS source,
    location_id,
    trace_number AS trace_key,
    COUNT(*) AS active_groups
FROM era_payment_groups
WHERE status NOT IN ('EXCEPTION', 'EXCEPTION_UNMATCHED')
GROUP BY location_id, trace_number
HAVING COUNT(*) > 1
UNION ALL
SELECT
    'active trace collisions' AS check_name,
    'expect zero rows; any row means more than one active group shares the same trace key' AS what_to_validate,
    'vcc' AS source,
    location_id,
    trace_id AS trace_key,
    COUNT(*) AS active_groups
FROM vcc_payment_groups
WHERE status NOT IN ('EXCEPTION', 'EXCEPTION_UNMATCHED')
GROUP BY location_id, trace_id
HAVING COUNT(*) > 1
ORDER BY source, location_id, trace_key;

-- 5) Match yield and backlog by source.
-- Returns one row for ERA groups and one row for VCC groups.
-- matched_groups is the main success metric.
-- awaiting_groups shows the active backlog.
-- exception_groups and expired_unmatched_groups show cases the matcher could not
-- safely resolve.
-- matched_rate_pct is a simple throughput ratio over all tracked groups.
SELECT
    'match yield and backlog' AS check_name,
    'matched_groups should trend up; awaiting_groups should stay bounded' AS what_to_validate,
    'era' AS source,
    COUNT(*) AS total_groups,
    COUNT(*) FILTER (WHERE status = 'MATCHED') AS matched_groups,
    COUNT(*) FILTER (WHERE status = 'AWAITING_VCC') AS awaiting_groups,
    COUNT(*) FILTER (WHERE status = 'EXCEPTION') AS exception_groups,
    COUNT(*) FILTER (WHERE status = 'EXCEPTION_UNMATCHED') AS expired_unmatched_groups,
    ROUND(
        100.0 * COUNT(*) FILTER (WHERE status = 'MATCHED')
        / NULLIF(COUNT(*), 0),
        2
    ) AS matched_rate_pct
FROM era_payment_groups
UNION ALL
SELECT
    'match yield and backlog' AS check_name,
    'matched_groups should trend up; awaiting_groups should stay bounded' AS what_to_validate,
    'vcc' AS source,
    COUNT(*) AS total_groups,
    COUNT(*) FILTER (WHERE status = 'MATCHED') AS matched_groups,
    COUNT(*) FILTER (WHERE status = 'AWAITING_ERA') AS awaiting_groups,
    COUNT(*) FILTER (WHERE status = 'EXCEPTION') AS exception_groups,
    COUNT(*) FILTER (WHERE status = 'EXCEPTION_UNMATCHED') AS expired_unmatched_groups,
    ROUND(
        100.0 * COUNT(*) FILTER (WHERE status = 'MATCHED')
        / NULLIF(COUNT(*), 0),
        2
    ) AS matched_rate_pct
FROM vcc_payment_groups;

-- 6) Match latency for successfully paired groups.
-- Returns the average and worst-case time from first receipt to match.
-- Lower numbers mean the system is reconciling more quickly once both sides exist.
SELECT
    'match latency' AS check_name,
    'avg_hours_to_match and max_hours_to_match should be low and stable' AS what_to_validate,
    'era' AS source,
    COUNT(*) AS matched_groups,
    ROUND(AVG(EXTRACT(EPOCH FROM (matched_at - first_received_at))) / 3600.0, 2) AS avg_hours_to_match,
    ROUND(MAX(EXTRACT(EPOCH FROM (matched_at - first_received_at))) / 3600.0, 2) AS max_hours_to_match
FROM era_payment_groups
WHERE status = 'MATCHED'
  AND matched_at IS NOT NULL
  AND first_received_at IS NOT NULL
UNION ALL
SELECT
        'match latency' AS check_name,
        'avg_hours_to_match and max_hours_to_match should be low and stable' AS what_to_validate,
    'vcc' AS source,
    COUNT(*) AS matched_groups,
    ROUND(AVG(EXTRACT(EPOCH FROM (matched_at - first_received_at))) / 3600.0, 2) AS avg_hours_to_match,
    ROUND(MAX(EXTRACT(EPOCH FROM (matched_at - first_received_at))) / 3600.0, 2) AS max_hours_to_match
FROM vcc_payment_groups
WHERE status = 'MATCHED'
  AND matched_at IS NOT NULL
  AND first_received_at IS NOT NULL;

-- 7) Reconciled-pair integrity checks.
-- Returns any reconciled payment whose linked groups disagree on amount or
-- provider identity signals. This should normally return zero rows.
-- Any row here is a likely false-positive or a data-quality edge case that should
-- be investigated manually.
SELECT
    'reconciled pair integrity' AS check_name,
    'expect zero rows; any row means the reconciled payment disagrees with its linked groups' AS what_to_validate,
    r.reconciled_payment_id,
    r.location_id,
    r.trace_number,
    r.status AS reconciled_status,
    r.matched_amount,
    e.bpr_amount AS era_amount,
    v.total_amount AS vcc_amount,
    r.provider_npi AS reconciled_provider_npi,
    v.provider_npi AS vcc_provider_npi,
    r.provider_tax_id AS reconciled_provider_tax_id,
    v.provider_tax_id AS vcc_provider_tax_id
FROM reconciled_payments r
JOIN era_payment_groups e ON e.group_id = r.era_payment_group_id
JOIN vcc_payment_groups v ON v.group_id = r.vcc_payment_group_id
WHERE r.matched_amount IS DISTINCT FROM e.bpr_amount
   OR r.matched_amount IS DISTINCT FROM v.total_amount
   OR r.provider_npi IS DISTINCT FROM v.provider_npi
   OR r.provider_tax_id IS DISTINCT FROM v.provider_tax_id
ORDER BY r.location_id, r.trace_number, r.reconciled_payment_id;

-- 8) State transition latency by entity.
-- Returns the average and worst observed time between consecutive transitions for
-- each entity type. Large values usually mean a queue, an external dependency, or
-- an operator handoff is slowing the pipeline down.
WITH ordered_transitions AS (
    SELECT
        entity_type,
        entity_id,
        from_state,
        to_state,
        transitioned_at,
        LAG(transitioned_at) OVER (
            PARTITION BY entity_type, entity_id
            ORDER BY transitioned_at
        ) AS previous_transitioned_at
    FROM state_transitions
)
SELECT
    'state transition latency' AS check_name,
    'avg_minutes_between_transitions and max_minutes_between_transitions should stay low' AS what_to_validate,
    entity_type,
    COUNT(*) FILTER (WHERE previous_transitioned_at IS NOT NULL) AS measured_hops,
    ROUND(
        AVG(EXTRACT(EPOCH FROM (transitioned_at - previous_transitioned_at))) / 60.0,
        2
    ) AS avg_minutes_between_transitions,
    ROUND(
        MAX(EXTRACT(EPOCH FROM (transitioned_at - previous_transitioned_at))) / 60.0,
        2
    ) AS max_minutes_between_transitions
FROM ordered_transitions
WHERE previous_transitioned_at IS NOT NULL
GROUP BY entity_type
ORDER BY entity_type;

-- 9) Stuck entities.
-- Returns the latest known state for any entity whose last transition is older
-- than the threshold below and is not in an obviously terminal state.
-- Adjust the interval as needed for your SLA; 24 hours is a practical starting point.
WITH latest_transition AS (
    SELECT DISTINCT ON (entity_type, entity_id)
        entity_type,
        entity_id,
        from_state,
        to_state,
        transitioned_at,
        reason
    FROM state_transitions
    ORDER BY entity_type, entity_id, transitioned_at DESC
)
SELECT
    'stuck entities' AS check_name,
    'expect zero rows; any row is an entity that has not progressed within the SLA window' AS what_to_validate,
    entity_type,
    entity_id,
    from_state,
    to_state,
    transitioned_at,
    reason,
    ROUND(EXTRACT(EPOCH FROM (NOW() - transitioned_at)) / 3600.0, 2) AS hours_since_last_transition
FROM latest_transition
WHERE transitioned_at < NOW() - INTERVAL '24 hours'
  AND to_state NOT IN ('MATCHED', 'PAYMENT_SUCCEEDED', 'PAYMENT_FAILED', 'POSTED', 'PARTIALLY_POSTED', 'WRITEBACK_FAILED', 'NOTIFIED', 'EXCEPTION', 'EXCEPTION_UNMATCHED')
ORDER BY transitioned_at ASC, entity_type, entity_id;

-- 10) Failure hotspots by transition target.
-- Returns counts of transitions into failure or exception-like states, grouped by
-- entity type and the reason text when present. This is the fastest way to see
-- where the system is losing reliability.
SELECT
    'failure hotspots' AS check_name,
    'look for large counts or repeated reasons in failure states' AS what_to_validate,
    entity_type,
    to_state,
    COALESCE(NULLIF(reason, ''), '(no reason)') AS reason_bucket,
    COUNT(*) AS transitions
FROM state_transitions
WHERE to_state IN (
    'EXCEPTION',
    'EXCEPTION_PARSE_FAILED',
    'EXCEPTION_UNMATCHED',
    'PROCESSING_FAILED',
    'PAYMENT_FAILED',
    'WRITEBACK_FAILED',
    'PARTIALLY_POSTED'
)
GROUP BY entity_type, to_state, COALESCE(NULLIF(reason, ''), '(no reason)')
ORDER BY transitions DESC, entity_type, to_state, reason_bucket;

-- 11) Suspicious or invalid transition patterns.
-- Returns events that look wrong from a state-machine perspective:
--   - a record transitioned to the same state twice in a row
--   - a transition has no prior state but is not an initial ingest state
--   - a transition jumps into a terminal state without a visible precursor
-- This is heuristic, but it is a good bug-finding screen when reliability looks off.
WITH ordered_transitions AS (
    SELECT
        entity_type,
        entity_id,
        from_state,
        to_state,
        transitioned_at,
        reason,
        LAG(to_state) OVER (
            PARTITION BY entity_type, entity_id
            ORDER BY transitioned_at
        ) AS previous_to_state
    FROM state_transitions
)
SELECT
    'suspicious transitions' AS check_name,
    'expect zero rows; any row suggests a state-machine bug or unusual replay' AS what_to_validate,
    entity_type,
    entity_id,
    from_state,
    to_state,
    transitioned_at,
    reason,
    CASE
        WHEN from_state = to_state THEN 'repeated same-state transition'
        WHEN from_state IS NULL AND to_state NOT IN ('RECEIVED_RAW', 'PARSED', 'AWAITING_ERA', 'AWAITING_VCC', 'MATCHED', 'PROCESSING_PAYMENT') THEN 'unexpected first transition'
        WHEN previous_to_state IS NOT NULL AND previous_to_state = to_state THEN 'duplicate consecutive state'
        ELSE 'suspicious'
    END AS suspicion
FROM ordered_transitions
WHERE from_state = to_state
   OR (from_state IS NULL AND to_state NOT IN ('RECEIVED_RAW', 'PARSED', 'AWAITING_ERA', 'AWAITING_VCC', 'MATCHED', 'PROCESSING_PAYMENT'))
   OR (previous_to_state IS NOT NULL AND previous_to_state = to_state)
ORDER BY transitioned_at DESC, entity_type, entity_id;