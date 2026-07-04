# Scenarios

These files are small scripts for checking how the Postgres functions behave.

They are not migrations or production scripts. They are meant to be run against a local throwaway database while building the ledger.

Each scenario starts with:

```sql
truncate table external_transfers, ledger_entries, ledger_transactions, ledger_accounts restart identity;
\ir ../migrations/003_seed_system_accounts.sql
```

That deletes all local scenario rows, resets generated IDs back to `1`, and then reinitializes system accounts like `Cash Settlement`.

That is useful here because each scenario needs a clean starting point. It makes the output easy to read and keeps the examples repeatable.

Do not run these against a real database. The `truncate` line will wipe the ledger data in those tables.

Run a scenario from the repo root:

```bash
psql "postgresql://ledger_db:password@localhost:5432/ledger_db" \
  -f projects/ledger-db/db/scenarios/001_alice_sends_bob.sql
```

Available scenarios:

- `001_alice_sends_bob.sql`: valid transfer
- `002_idempotency.sql`: same request returns the original transaction
- `003_insufficient_funds.sql`: transfer fails before moving money
- `004_mismatched_idempotency_key.sql`: same key with different request fields fails
- `005_compare_stored_and_derived_balances.sql`: stored balances match balances derived from entries
