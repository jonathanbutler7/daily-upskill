# Scenarios

These files are small scripts for checking how the Postgres function `post_transfer` behaves.

They are not migrations or production scripts. They are meant to be run against a local throwaway database while building the ledger.

Each scenario starts with:

```sql
truncate table ledger_entries, ledger_transactions, ledger_accounts restart identity;
```

That deletes all rows from the ledger tables and resets the generated IDs back to `1`.

That is useful here because each scenario needs a clean starting point. It makes the output easy to read and keeps the examples repeatable.

Do not run these against a real database. The `truncate` line will wipe the ledger data in those tables.

Run a scenario from the repo root:

```bash
psql "postgresql://ledger_db:password@localhost:5432/ledger_db" \
  -f projects/ledger-db/scenarios/001_alice_sends_bob.sql
```

Available scenarios:

- `001_alice_sends_bob.sql`: valid transfer
- `002_idempotency.sql`: same request returns the original transaction
- `003_insufficient_funds.sql`: transfer fails before moving money
- `004_mismatched_idempotency_key.sql`: same key with different request fields fails
