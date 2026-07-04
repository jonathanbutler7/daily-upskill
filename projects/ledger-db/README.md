# Ledger DB

See [success criteria](docs/success-criteria.md) for the end state of this project.

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
  -f projects/ledger-db/db/migrations/001_create_ledger_tables.sql

psql "postgresql://ledger_db:password@localhost:5432/ledger_db" \
  -f projects/ledger-db/db/migrations/002_create_external_transfers.sql

psql "postgresql://ledger_db:password@localhost:5432/ledger_db" \
  -f projects/ledger-db/db/migrations/003_seed_system_accounts.sql

psql "postgresql://ledger_db:password@localhost:5432/ledger_db" \
  -f projects/ledger-db/db/migrations/004_create_post_transfer_function.sql

psql "postgresql://ledger_db:password@localhost:5432/ledger_db" \
  -f projects/ledger-db/db/migrations/005_create_add_balance_function.sql
```

```sql
\i /Users/jonathanbutler/projects/daily-upskill/projects/ledger-db/db/migrations/004_create_post_transfer_function.sql
```

## Test `post_transfer`

Load the simple Alice/Bob scenario:

```bash
psql "postgresql://ledger_db:password@localhost:5432/ledger_db" \
  -f projects/ledger-db/db/scenarios/001_alice_sends_bob.sql
```

Or call the function manually inside `psql`:

```sql
select add_balance(2, 2000, 'ach', 'manual-deposit-ext-1', 'manual-deposit-1');
select post_transfer(2, 3, 1000, 'test-1');

select * from ledger_accounts;
select * from ledger_transactions order by id;
select * from ledger_entries order by id;
```

What this has proved so far:

- Empty tables fail with `from account not found`.
- A deposit posts balanced entries against `Cash Settlement`.
- A valid transfer moves balance from one account to another.
- Ledger entries are created with equal and opposite amounts.
- Insufficient funds fails before moving money.
- Reusing the same idempotency key with the same request returns the original transaction.
- Reusing the same idempotency key with a different request fails.

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
