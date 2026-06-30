# Data Model

This document describes the target database integration model for PayerSync.

## Tables

| # | Table | Description | Status |
|---|-------|-------------|--------|
| 1 | `era_remittances` | Parsed ERA 835 file | implemented |
| 2 | `era_payment_groups` | Payment-level unit per ERA | implemented |
| 3 | `vcc_files` | VCC file metadata | implemented |
| 4 | `vcc_rows` | Individual CSV records | implemented |
| 5 | `vcc_payment_groups` | Grouped VCC rows, the unit matched by the reconciler | implemented |
| 6 | `reconciled_payments` | Joined ERA + VCC payment group | planned |
| 7 | `ledger_postings` | PMS writeback attempt and result | planned |
| 8 | `job_runs` | Cron job execution log | implemented |
| 9 | `state_transitions` | Append-only state transition log | implemented |

---

## Implemented now

These tables exist in `db/migrations` today and are the current database source of truth.

## Audit report

For a reusable Postgres audit of deduping and matching reliability, see [audit_report.sql](audit_report.sql).

## Parsed Data

### `era_remittances`

**Status:** implemented

| Field | Notes |
|-------|-------|
| `era_id` | PK |
| `location_id` | |
| `payer_name` | |
| `provider_npi` | |
| `provider_tax_id` | |
| `bpr_amount` | |
| `status` | `RECEIVED_RAW` \| `PARSED` \| `EXCEPTION_PARSE_FAILED` |
| `payment_method` | |
| `trace_number` | |
| `received_at` | |
| `file_hash` | used for dedup |
| `raw_storage_key` | bucket reference |

### `era_payment_groups`

**Status:** implemented

| Field | Notes |
|-------|-------|
| `era_payment_group_id` | PK |
| `era_id` | FK → era_remittances |
| `location_id` | |
| `trace_number` | primary match key |
| `bpr_amount` | |
| `claim_count` | |
| `claims` | json |
| `adjustments` | json |
| `status` | `AWAITING_VCC` \| `MATCHED` \| `EXCEPTION` |
| `reconciliation_triggered_at` | idempotency guard for downstream trigger |

### `vcc_files`

**Status:** implemented

| Field | Notes |
|-------|-------|
| `vcc_file_id` | PK |
| `location_id` | |
| `received_at` | |
| `file_hash` | used for dedup |
| `raw_storage_key` | bucket reference |
| `row_count` | |
| `source_filename` | |
| `status` | `RECEIVED_RAW` \| `PARSED` \| `EXCEPTION_PARSE_FAILED` |

### `vcc_rows`

**Status:** implemented

| Field | Notes |
|-------|-------|
| `vcc_row_id` | PK |
| `vcc_file_id` | FK → vcc_files |
| `vcc_payment_group_id` | FK → vcc_payment_groups |
| `payment_id` | |
| `trace_id` | primary match key |
| `payer_name` | |
| `provider_npi` | |
| `provider_tax_id` | |
| `issue_date` | |
| `amount` | |
| `card_fingerprint` | |
| `last4` | |
| `expiration_date` | |
| `patient_id` | |
| `claim_id` | |
| `service_date_start` | |
| `service_date_end` | |

### `vcc_payment_groups`

**Status:** implemented

| Field | Notes |
|-------|-------|
| `vcc_payment_group_id` | PK |
| `trace_id` | primary match key |
| `payment_id` | |
| `provider_npi` | |
| `provider_tax_id` | |
| `card_fingerprint` | |
| `total_amount` | sum of member row amounts |
| `location_id` | |
| `status` | `AWAITING_ERA` \| `MATCHED` \| `EXCEPTION` |
| `reconciliation_triggered_at` | idempotency guard for downstream trigger |

---

## Payment Data

### `reconciled_payments`

**Status:** planned

| Field | Notes |
|-------|-------|
| `reconciled_payment_id` | PK |
| `era_payment_group_id` | FK → era_payment_groups |
| `vcc_payment_group_id` | FK → vcc_payment_groups |
| `amount` | |
| `status` | `PROCESSING_PAYMENT` \| `PAYMENT_SUCCEEDED` \| `PROCESSING_FAILED` \| `WRITING_BACK` \| `POSTED` \| `PARTIALLY_POSTED` \| `WRITEBACK_FAILED` |
| `matched_at` | |
| `processed_at` | |
| `attempted_at` | |
| `error` | |
| `retries` | |

### `ledger_postings`

**Status:** planned

| Field | Notes |
|-------|-------|
| `ledger_posting_id` | PK |
| `reconciled_payment_id` | FK → reconciled_payments |
| `idempotency_key` | |
| `pms` | target practice management system |
| `status` | `PENDING` \| `SUCCESS` \| `FAILED` |
| `attempted_at` | |
| `response` | json |
| `error` | nullable |

---

## Audit Logs

### `job_runs`

**Status:** implemented

| Field | Notes |
|-------|-------|
| `run_id` | PK |
| `job_type` | `ingester` \| `reconciler` \| `processor` \| `writeback` |
| `started_at` | |
| `finished_at` | |
| `status` | `success` \| `failure` \| `partial` |
| `files_processed` | |
| `records_matched` | |
| `errors` | json |

### `state_transitions`

**Status:** implemented

| Field | Notes |
|-------|-------|
| `transition_id` | PK |
| `entity_type` | `era` \| `vcc` \| `era_remittance` \| `vcc_file` \| `era_payment_group` \| `vcc_payment_group` \| `reconciled_payment` \| `ledger_posting` |
| `entity_id` | |
| `from_state` | |
| `to_state` | |
| `transitioned_at` | |
| `reason` | nullable |
