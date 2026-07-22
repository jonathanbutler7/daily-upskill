# Reconciliation Engine TPD

## Summary

`reconciliation-engine` compares external payment reports against internal `ledger-db` state and creates durable match results and exceptions.

The first version should be a small service with its own Postgres tables. It should import CSV or JSON settlement report fixtures, read `ledger-db` external transfers, run deterministic matching, and record exceptions. It should not mutate ledger rows directly.

For the MVP, assume `ledger-db` already exists and `rail-sim` does not. Imported report files are the external source of truth. Later, `rail-sim` can become another report source that produces the same normalized report shape.

## Product Boundary

`reconciliation-engine` owns:

- report imports
- normalized report rows
- reconciliation runs
- match results
- reconciliation exceptions
- exception resolution status

`ledger-db` owns:

- ledger accounts
- ledger transactions
- ledger entries
- balances
- external transfer records
- idempotency
- reversals and true-up commands

Future external sources can own:

- external rail events
- settlement report generation
- processor or bank report exports
- returned payment events

`rail-sim` is a future adapter, not an MVP dependency.

## Architecture

For the MVP, build a single service with a relational database.

```text
CSV or JSON report fixture
        |
        v
POST /reports
        |
        v
report_imports + report_rows
        |
        v
POST /reports/{id}/run
        |
        v
read ledger-db external_transfers
        |
        v
match_results + reconciliation_exceptions
```

The matching engine should be deterministic. The same report and same ledger state should produce the same results.

## MVP Assumptions

- `ledger-db` exists and has external transfer records with stable external references.
- `rail-sim` does not exist yet.
- External truth arrives as an uploaded or locally loaded CSV/JSON report.
- The report format is normalized during import so future sources can reuse the same matching logic.
- Reconciliation reads from `ledger-db` but does not update ledger tables directly.

## Data Model

### `reconciliation_reports`

Stores one imported report.

```text
id
source
rail
report_date
file_hash
status
imported_at
completed_at
```

Unique key:

```text
source + rail + report_date + file_hash
```

This prevents duplicate report imports.

### `reconciliation_report_rows`

Stores normalized external report rows.

```text
id
report_id
external_reference
amount
currency_code
direction
external_status
settled_at
raw_row
created_at
```

`raw_row` should keep the original normalized payload for debugging.

### `reconciliation_runs`

Stores each matching attempt.

```text
id
report_id
status
started_at
completed_at
matched_count
exception_count
```

The MVP can allow reruns. A rerun should create a new run record and replace or supersede old open match results for that report.

### `match_results`

Stores the result for each report row.

```text
id
run_id
report_row_id
external_reference
ledger_external_transfer_id
result
details
created_at
```

Allowed results:

```text
matched
missing_in_ledger
missing_in_report
amount_mismatch
direction_mismatch
status_mismatch
date_mismatch
duplicate_reference
```

### `reconciliation_exceptions`

Stores mismatches that need review or resolution.

```text
id
run_id
match_result_id
external_reference
reason
status
resolution_type
ledger_transaction_id
notes
created_at
resolved_at
```

Allowed statuses:

```text
open
in_review
resolved
ignored
```

Allowed resolution types:

```text
none
ledger_reversal_needed
ledger_true_up_needed
external_report_error
duplicate_report_row
no_action_needed
```

## Matching Rules

The first version should match by `external_reference`.

For each report row:

1. Find matching `ledger-db.external_transfers.external_reference`.
2. If no ledger record exists, create `missing_in_ledger`.
3. If more than one report row uses the same reference, create `duplicate_reference`.
4. If the amount differs, create `amount_mismatch`.
5. If the direction differs, create `direction_mismatch`.
6. If the external status conflicts with ledger external transfer status, create `status_mismatch`.
7. If the settlement date is outside the expected window, create `date_mismatch`.
8. If all required fields match, create `matched`.

After report-row matching, scan `ledger-db` external transfers for the report date and rail. Any ledger transfer that should appear in the report but does not should create `missing_in_report`.

## Ledger DB Read Contract

The reconciliation service needs a stable read model from `ledger-db`.

Minimum fields:

```text
external_transfer_id
external_reference
ledger_transaction_id
user_account_id
direction
amount
currency_code
status
created_at
settled_at
```

For the MVP, this can be a SQL query against the local `ledger-db` database. Later it can become a `ledger-db` HTTP API.

The reconciliation service should never update `ledger-db` tables directly.

## Resolution Flow

Resolving an exception should record the decision in `reconciliation-engine`.

Examples:

- `external_report_error`: the external report row was wrong.
- `duplicate_report_row`: the report had a duplicate row and no ledger change is needed.
- `ledger_reversal_needed`: a posted ledger movement needs a reversal.
- `ledger_true_up_needed`: a manual correction is needed.
- `no_action_needed`: the mismatch is understood and accepted.

In the MVP, resolution should not automatically call `ledger-db`. The resolution record can store the `ledger_transaction_id` if a user later performs a reversal or true-up.

## API

```text
POST /reports
GET /reports/{report_id}
POST /reports/{report_id}/run
GET /reports/{report_id}/summary
GET /reports/{report_id}/matches
GET /exceptions
GET /exceptions/{exception_id}
POST /exceptions/{exception_id}/resolve
```

### `POST /reports`

Imports a report.

For the MVP, expected sources are:

```text
csv_fixture
json_fixture
manual_import
```

Future sources can include:

```text
rail_sim
processor_report
bank_report
```

Input:

```json
{
  "source": "csv_fixture",
  "rail": "ach",
  "report_date": "2026-07-21",
  "rows": [
    {
      "external_reference": "ach_001",
      "amount": 200000,
      "currency_code": "USD",
      "direction": "deposit",
      "external_status": "settled",
      "settled_at": "2026-07-21T14:00:00Z"
    }
  ]
}
```

Behavior:

- compute a stable `file_hash` from the normalized payload
- reject invalid rows
- return the existing report if the same report was already imported

For CSV input, the MVP should support this shape:

```csv
external_reference,amount,currency_code,direction,external_status,settled_at
ach_001,200000,USD,deposit,settled,2026-07-21T14:00:00Z
```

### `POST /reports/{report_id}/run`

Runs reconciliation for one report.

Behavior:

- load report rows
- load matching `ledger-db` external transfers
- create match results
- create exceptions for non-matches
- return a summary

### `POST /exceptions/{exception_id}/resolve`

Records a resolution decision.

Input:

```json
{
  "resolution_type": "ledger_reversal_needed",
  "ledger_transaction_id": "txn_123",
  "notes": "Returned after the wallet was credited."
}
```

## Failure Handling

- Invalid report rows should fail import before any rows are stored.
- Duplicate imports should return the original report.
- A failed reconciliation run should keep its run record with `failed` status.
- Match creation should happen in one database transaction per run.
- A rerun should not create duplicate open exceptions for the same report, reference, and reason.

## Testing Plan

Use DB-backed tests for the matching behavior because the important guarantees come from persistence, uniqueness, and replay safety.

Test cases:

- import a valid JSON report
- import a valid CSV report
- reject an invalid report row
- import the same report twice without duplicating rows
- match a settled report row to a `ledger-db` external transfer
- create `missing_in_ledger`
- create `missing_in_report`
- create `amount_mismatch`
- create `duplicate_reference`
- resolve an exception without editing ledger data
- rerun a report without duplicate active exceptions

## Implementation Order

1. Define report, row, run, match, and exception tables.
2. Build JSON report import with validation and idempotency.
3. Build CSV report import into the same normalized row model.
4. Build the `ledger-db` read query.
5. Build matching for exact reference matches.
6. Add mismatch detection.
7. Add exception creation.
8. Add summary endpoint.
9. Add exception resolution.
10. Add DB-backed tests for replay and mismatch cases.

## Open Questions

- Should the MVP share the same Postgres instance as `ledger-db` or use a separate database?
- Should `missing_in_report` use report date, settlement date, or created date to choose eligible ledger transfers?
- Which `ledger-db` external transfer statuses should be treated as report-eligible?
- Should resolved exceptions be immutable, or can notes be edited?
- When should resolution call `ledger-db` directly instead of only recording a decision?
- What normalized fields should future `rail-sim` reports provide so they can use the same import path?
