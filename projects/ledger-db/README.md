# Ledger DB

## Local setup

This project is using a local Homebrew Postgres install.

Connect as the admin user through the default `postgres` database:

```bash
psql postgres
```

Create the local exercise role and database:

```sql
CREATE ROLE ledger_db WITH LOGIN PASSWORD 'password';
CREATE DATABASE ledger_db OWNER ledger_db;
```

Connect as the app role:

```bash
psql "postgresql://ledger_db:password@localhost:5432/ledger_db"
```

## Run migrations

Run these from the repo root:

```bash
psql "postgresql://ledger_db:password@localhost:5432/ledger_db" \
  -f ledger-db/migrations/001_create_ledger_tables.sql

psql "postgresql://ledger_db:password@localhost:5432/ledger_db" \
  -f ledger-db/migrations/002_create_post_transfer_function.sql
```

```sql
\i /Users/jonathanbutler/projects/daily-upskill/ledger-db/migrations/002_create_post_transfer_function.sql
```

## Test `post_transfer`

Load the simple Alice/Bob scenario:

```bash
psql "postgresql://ledger_db:password@localhost:5432/ledger_db" \
  -f ledger-db/scenarios/001_alice_sends_bob.sql
```

Or call the function manually inside `psql`:

```sql
select post_transfer(1, 2, 1000, 'test-1');

select * from ledger_accounts;
select * from ledger_transactions order by id;
select * from ledger_entries order by id;
```

What this has proved so far:

- Empty tables fail with `from account not found`.
- A valid transfer moves balance from one account to another.
- Ledger entries are created with equal and opposite amounts.
- Insufficient funds fails before moving money.
- Reusing the same idempotency key currently fails on the unique constraint. The next step is to return the original transaction instead of erroring.

## Boundaries

- Postgres owns 
  - validation of current db state before initiating transfer
  - locking
  - atomic write
  - idempotency constraint

- Application code owns 
  - API shape
  - request parsing
  - retries
  - error mapping
  - tests
