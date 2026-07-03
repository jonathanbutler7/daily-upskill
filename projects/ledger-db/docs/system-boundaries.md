# System Boundaries

This project uses both Postgres and Go deliberately.

The rule is:

- Put durable ledger guarantees in Postgres.
- Put command shape, workflow, retries, and caller-facing behavior in Go.
- Do not rely on Go for checks that can go stale between a read and a write.
- Do not hide normal application behavior inside Postgres when Go can make it clearer.

## Postgres owns

Postgres owns the facts that must stay true no matter which application path writes to the ledger.

- Ledger tables and relationships.
- Foreign keys between transactions, entries, and accounts.
- Unique idempotency keys.
- Basic checks like nonzero entry amounts and positive request amounts.
- Atomic writes through database transactions.
- Row locks for accounts whose balances are being changed.
- Final balance checks before money leaves an account.
- Stored balances and the entries used to derive them.
- Protection against corrupted posted ledger state.

The important point: Go can check some of these early, but Postgres has to be the final guard.

Example: Go can reject `amount <= 0` before touching the database. Postgres should still reject it because another caller or a future code path could bypass that Go check.

Example: Go should not decide that Alice has enough money based only on an earlier read. Postgres needs to check the current locked account row at posting time.

## Go owns

Go owns the shape of the system that callers use.

- HTTP or CLI command shape.
- Request parsing.
- Request-level validation.
- Idempotency behavior from the caller's point of view.
- Mapping database errors to stable application errors.
- Retry behavior for serialization failures, worker retries, and timeouts.
- Multi-step workflows that involve more than one ledger operation.
- External payment integration flow.
- Reconciliation jobs.
- Tests around the application boundary.

The important point: Go should make the system usable and extensible. It should not just be a thin wrapper that leaks raw database behavior to callers.

Example: the database might raise `insufficient funds`. Go should return a clear application error instead of exposing a raw SQL error string.

Example: a duplicate idempotency key with the same request should return the original transaction to the caller. A duplicate key with different request fields should become a conflict error.

## Shared responsibilities

Some rules belong in both places, but for different reasons.

| Rule | Go responsibility | Postgres responsibility |
| --- | --- | --- |
| Positive transfer amount | Reject bad requests early | Enforce the rule even if Go is bypassed |
| Idempotency | Define caller behavior and return shape | Store unique key and prevent double posting |
| Insufficient funds | Return a clear application error | Check the locked balance before posting |
| Currency match | Validate request shape and report cleanly | Prevent cross-currency posting without an FX flow |
| Balanced entries | Build the intended entries | Reject posted transactions that do not balance |
| Reversals | Expose a reversal command and error model | Preserve original entries and prevent double reversal |

## Current project boundary

The project currently has most posting behavior in Postgres functions because that was the easiest way to prove the ledger rules while learning the schema.

That is not the final shape.

The next version should keep the database invariants, but introduce a small Go boundary around the existing operations:

- `Transfer`
- `Deposit`
- later, `ReverseTransaction`

The Go layer should call into database code that performs the atomic write. Go should own the command types, request validation, error mapping, retries, and tests.

## Next decision

Before adding a large Go service, define the command boundary:

```text
Transfer(from_account_id, to_account_id, amount, currency_code, idempotency_key)
Deposit(to_account_id, amount, currency_code, external_reference, idempotency_key)
ReverseTransaction(original_transaction_id, reason, idempotency_key)
```

Then decide which parts are stable database invariants and which parts are application workflow.

That gives the project room to grow while keeping each rule in the layer that can enforce it best.
