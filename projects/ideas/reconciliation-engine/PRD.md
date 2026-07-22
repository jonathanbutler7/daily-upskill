# Reconciliation Engine PRD

## Summary

Build `reconciliation-engine`, a small payments reconciliation product that compares internal ledger state against external payment reports and identifies mismatches.

The project should integrate with `ledger-db` and `rail-sim`. `ledger-db` owns internal accounting. `rail-sim` can provide external settlement reports and rail events. `reconciliation-engine` owns matching, exception detection, review status, and resolution workflow.

This is a core payments system because money movement is only trustworthy when internal records agree with external evidence.

## Problem

`ledger-db` can post balanced transactions and store external transfer metadata, and `rail-sim` can produce external rail events. The missing product is the system that compares those sources and answers:

- did every settled external payment get posted?
- did every posted external transfer actually settle?
- do the amounts, dates, directions, and references match?
- which differences need manual review?
- which differences require a ledger reversal, adjustment, or true-up?

Without this layer, mismatches stay hidden until a user balance, cash account, or settlement report looks wrong.

## Goals

- Ingest external settlement reports from `rail-sim` or CSV fixtures.
- Load matching internal records from `ledger-db`.
- Match report rows to `ledger-db` external transfers by stable reference.
- Detect missing, duplicate, amount mismatch, direction mismatch, status mismatch, and date mismatch cases.
- Create reconciliation exceptions with clear reasons.
- Track exception status from open to resolved.
- Record resolution decisions without editing original ledger history.
- Support replaying the same report safely.
- Produce a simple reconciliation summary by report date.

## Non-Goals

- Real bank, ACH, card, or processor integrations.
- Machine-learning matching.
- Full finance close workflow.
- General accounting software replacement.
- Automatic money movement.
- Automatic ledger corrections in the MVP.
- Complex UI beyond a simple admin/API review surface.

## Users

### Primary User

An operations or finance user responsible for daily payments review.

They need to know which payments matched, which did not, and what needs action.

### Secondary User

The developer building payments systems knowledge.

They need a project that connects ledger accounting, external rail state, idempotent ingestion, and exception handling.

### Reviewer

An interviewer or technical reviewer.

They should be able to see that the project models a realistic payments control system instead of only happy-path transfers.

## Core Concepts

### Reconciliation Report

An imported external source for a specific date, rail, or provider.

Example fields:

- `report_id`
- `source`
- `report_date`
- `file_hash`
- `status`
- `imported_at`

### Report Row

A normalized row from an external settlement report.

Example fields:

- `report_row_id`
- `external_reference`
- `amount`
- `currency_code`
- `direction`
- `external_status`
- `settled_at`

### Match Result

The comparison between a report row and internal `ledger-db` state.

Possible results:

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

### Reconciliation Exception

A durable record that explains a mismatch and tracks how it was resolved.

Example fields:

- `exception_id`
- `report_id`
- `external_reference`
- `reason`
- `status`
- `resolution_type`
- `ledger_transaction_id`
- `notes`

## MVP Product Surface

`reconciliation-engine` should expose an HTTP API and store reconciliation state in Postgres.

```text
POST /reports
GET /reports/{report_id}
POST /reports/{report_id}/run
GET /reports/{report_id}/summary
GET /exceptions
GET /exceptions/{exception_id}
POST /exceptions/{exception_id}/resolve
```

`POST /reports` should accept a report payload or file reference. Re-importing the same report should be idempotent by `source`, `report_date`, and `file_hash`.

## Ledger DB Integration

The clean boundary:

- `ledger-db` owns ledger accounts, ledger transactions, ledger entries, balances, idempotency, reversals, and external transfers.
- `reconciliation-engine` reads `ledger-db` state and creates reconciliation records.
- `reconciliation-engine` should not edit ledger rows.
- Corrections should happen through explicit `ledger-db` reversal, adjustment, or true-up commands.

Expected matching behavior:

- Settled report row with matching posted ledger transfer: mark matched.
- Settled report row missing from `ledger-db`: create `missing_in_ledger`.
- Posted `ledger-db` external transfer missing from the report: create `missing_in_report`.
- Same reference with different amount: create `amount_mismatch`.
- Same reference with different direction: create `direction_mismatch`.
- Returned external event after posted ledger movement: create an exception that points to reversal or true-up.

## Rail Sim Integration

`rail-sim` can provide the external truth source for the MVP.

Example report row:

```json
{
  "external_reference": "ach_001",
  "rail": "ach",
  "direction": "deposit",
  "status": "settled",
  "amount": 200000,
  "currency_code": "USD",
  "settled_at": "2026-07-21T14:00:00Z"
}
```

`reconciliation-engine` should also support CSV fixtures so the product can be tested without running `rail-sim`.

## MVP Scenarios

1. A settled report row matches a `ledger-db` external transfer.
2. A settled report row is missing from `ledger-db`.
3. A `ledger-db` external transfer is missing from the report.
4. A report row and ledger record have the same reference but different amounts.
5. A duplicate report row is detected.
6. A returned payment creates an exception that requires reversal review.
7. Re-importing the same report does not create duplicate rows or exceptions.

## Success Criteria

- A developer can import a settlement report fixture.
- The system can run reconciliation against `ledger-db`.
- The system creates clear match results for every report row.
- The system creates durable exceptions for mismatches.
- The same report can be replayed safely.
- At least one matched case, one missing case, one amount mismatch, and one duplicate case are covered by tests.
- A reversal or true-up decision can be recorded without editing original ledger history.

## Open Questions

- Should reconciliation state live in its own database or inside `ledger-db` first?
- Should matching read directly from `ledger-db` tables or call a `ledger-db` API?
- Should the first version reconcile only external transfers, or also derived balances?
- Should exception resolution trigger `ledger-db` commands or only record the decision?
- Should reports be grouped by rail, provider, settlement date, or import batch?
