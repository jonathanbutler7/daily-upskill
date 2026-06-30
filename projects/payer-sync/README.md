# PayerSync

## Problem

Healthcare offices often receive insurance reimbursements in a fragmented way:

- remittance data arrives as **EDI X12 835 / ERA**  
- card funding details arrive separately as a **VCC CSV**  
- the two files may not arrive at the same time  
- one VCC payload may cover multiple claims  
- one office payment may need to be allocated across multiple claims or service lines  
- payment posting and office communication are often manual

The result is a slow, error-prone workflow with weak observability and poor reconciliation.

The prototype PayerSync product should reduce manual work by turning these asynchronous files into a single, traceable processing pipeline.

## Module Overview

Payer sync is composed of 5 distinct modules.

1. **Ingest** new ERA and VCC packet files by polling a remote server
2. **Reconcile** the two data sources and determine if a match exists
3. **Process** the payment using the VCC once a matching ERA is found
4. **Writeback** successfully processed charges to the user's PMS
5. **Notify** the user of successful or partially successful payments, with separate handling for failure notifications

## Invariants

- The system expects to receive ERA and VCC files from a remote server
- The system will provide an auditable trail for every stage of the process, including state transitions, when payments are processed, and when they are written back
- Because the upstream API is scoped to one office per credential, the sync job will run for each office in a list of office IDs, rather than one global sync across all offices

## Lifecycle

These statuses represent the possible statuses throughout the lifecycle of payer sync:

- `RECEIVED_RAW`
- `PARSED`
- `AWAITING_ERA`/`AWAITING_VCC`
- `MATCHED`
- `PROCESSING_PAYMENT`
- `PAYMENT_SUCCEEDED`
- `PAYMENT_FAILED`
- `WRITING_BACK`
- `POSTED`
- `PARTIALLY_POSTED`
- `WRITEBACK_FAILED`
- `NOTIFIED`
- `EXCEPTION`
- `ARCHIVED`

## Out of Scope

For an MVP version, the following items are out of scope

- UI to allow for operations to view failures and attempt retries
- UI that gives visibility into following a particular record through the pipeline

## Database Tooling

This project uses:

- Goose for migrations
- SQLC for type-safe query generation

### Connection Configuration

DB connections are environment-driven so switching between local Postgres and Supabase is low-friction.

1. Copy `.env.example` values into your environment.
2. Prefer `DATABASE_URL` when available.
3. If `DATABASE_URL` is unset, the app builds a DSN from `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, and `DB_NAME`.
4. Use `DB_SSLMODE=disable` for local and `DB_SSLMODE=require` for Supabase.

### Run Migrations

The ingester now applies any pending migrations automatically on startup before it begins processing files. These commands are still useful when you want to inspect or manage schema changes explicitly.

```bash
go run . migrate up
go run . migrate status
go run . migrate down
```

Or use Make targets:

```bash
make migrate-up
make migrate-status
make migrate-down
make audit-report
```

`make audit-report` runs the reusable Postgres reliability audit in [db/docs/audit_report.sql](db/docs/audit_report.sql) against the database configured in your environment.

### Running the Ingester

```bash
go run .        # apply pending migrations, then ingest whatever files the seeder has
```

### Private Key Handling

- If `INGESTER_PRIVATE_KEY` is set, the app uses that PEM directly and never writes a local key file.
- If `INGESTER_PRIVATE_KEY_PATH` is set, the app reads or creates the key at that absolute path.
- If neither is set, the app reads or creates the key under your user config directory at `payer-sync/ingester_private_key.pem` (for example `~/Library/Application Support/payer-sync/ingester_private_key.pem` on macOS).

For local development, the seeder starts with a fixed seed batch. Once those files have been ingested they become duplicates and subsequent runs will skip them. Use `seed` to reset both the seeder and the database and generate a fresh batch:

```bash
go run . seed   # reset seeder + DB, generate fresh files, ingest
```

> In production the seeder represents a real payer system whose files arrive on their own schedule. The ingester just polls; `seed` is a local-dev convenience only.

### Generate SQLC Code

```bash
make sqlc-generate
```

Generated code is written to `internal/db/store`.
