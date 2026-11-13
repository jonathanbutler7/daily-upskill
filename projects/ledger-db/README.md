# Ledger DB

I built this project to learn double-entry bookkeeping and how a ledger records money movement. 

Stack:
- Go
- Postgres

## Focus 

Keeping transfers correct and understanding how the database performs under concurrent load.

## What it does

Transfers between 
1. accounts
2. deposits
3. withdrawals
4. reversals
5. idempotency

## Tables

- external_transfers
- ledger_accounts
- ledger_transactions
- ledger_entries
- ledger_reversals

one ledger transaction -> two balanced entries. 

The balance checks, entries, and balance updates happen inside an atomic database transaction. If any step fails, the whole operation rolls back.


## The hot-account problem

Deposits and withdrawals use a shared Cash Settlement account as the other side of the ledger entry. 

In the synchronous path, every deposit and withdrawal updates that settlement account's balance. 

Postgres locks the row during the update, so concurrent operations have to wait for one another. Those operations hold database connections while waiting, which can also cause other requests to queue for an available connection.

The ledger stays correct, but throughput drops and requests take longer.

## Sync vs deferred async

Both paths use atomic database transactions. 

- Sync: update the settlement balance in one transaction
- Deferred async: add an outbox event in one transaction. A worker updates the settlement balance afterward.

The worker commits the settlement balance update and marks the events processed in the same transaction. If that transaction rolls back, the events stay pending for retry.

### Batching

The worker reads up to 100 pending events at a time. It sums their changes and updates each settlement account once per batch. 

### Backlog tradeoff

The customer balance and ledger entries are current when the request commits. The stored settlement balance can lag behind them.

If the worker stops or can't keep up, pending events accumulate. We need to monitor how many are pending and how old the oldest event is. The settlement balance catches up as the worker processes the backlog.

## Testing under load

Good run

```bash
go run ./cmd/stress -accounts=5 -workers=2 -operations=50
```

Bad run

```bash
go run ./cmd/stress -accounts=10 -workers=200 -operations=50
```

```bash
LEDGER_DB_DSN='postgresql://ledger_db:password@localhost:5432/ledger_db_interview_stress' \
go run ./cmd/stress \
  -ab-test \
  -accounts=500 \
  -workers=150 \
  -operations=10000 \
  -max-open-conns=50 \
  -external-transfer-percent=90 \
  -think-min=1ms \
  -think-max=1ms
```
