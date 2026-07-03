# Success Criteria

The goal of this project is to build a serious double-entry ledger system.

This is still a learning project, but it should not be a toy ledger. By the end, it should show the core data model, the money-movement rules, the service boundaries, the failure cases, and the production features a real ledger would need even if this project does not implement all of them.

## Done Means

The project is done when it can post, reject, reverse, and reconcile ledger activity through clear application and database boundaries, and when the docs explain which guarantees are implemented versus only recognized as production requirements.

This project should be strong enough to talk through as an industry-style ledger POC.

## Core Ledger Requirements

- [x] The schema has ledger accounts, ledger transactions, and ledger entries.
- [x] Ledger entries are the source of the accounting record.
- [x] Stored account balances exist as fast operational state.
- [x] Ledger accounts start at `0`; balances are created only by posted ledger activity.
- [x] Money movement happens through controlled posting logic, not ad hoc inserts.
- [x] A valid transfer creates one ledger transaction and two opposite entries.
- [x] A valid transfer updates stored balances in the same database transaction.
- [x] Insufficient funds fails without creating entries or changing balances.
- [x] Same idempotency key with the same request returns the original transaction.
- [x] Same idempotency key with a different request fails.
- [x] Concurrent retries with the same request move money once.
- [x] Stored balances can be compared against balances derived from entries.
- [ ] Each posted ledger transaction must balance to zero.
- [ ] The project must prove how unbalanced transactions are prevented.
- [ ] Posted ledger entries must be immutable.
- [ ] Posted ledger transactions must be immutable except for safe metadata.
- [ ] Direct table writes must be blocked, avoided through permissions, or explicitly called out as a POC limitation.

## External Money Movement Requirements

Money can enter or leave this ledger, but the ledger still has to stay balanced. A real ledger does not add money with a one-sided balance update.

- [x] The project recognizes that ledger accounts should start at `0`.
- [x] The project has a temporary `External Funding` account for opening balances and deposits.
- [ ] Replace the `External Funding` shortcut with a clearer external-money model.
- [ ] External bank accounts, card networks, processors, or payment rails should not be treated as normal user ledger accounts.
- [ ] The ledger should use an internal account, such as cash, settlement, clearing, external funds, or processor receivable, to represent money controlled or expected by the system.
- [ ] A deposit posts balanced entries between the receiving user account and the internal funding or settlement account.
- [ ] A withdrawal posts balanced entries between the sending user account and the internal funding or settlement account.
- [ ] External payment identifiers should be stored separately from ledger entries.
- [ ] External payment status should be reconciled against ledger activity.
- [ ] The docs explain the difference between an external real-world account and an internal ledger account that represents that external relationship.
- [ ] The docs explain whether deposits are posted only after settlement or first posted as pending/clearing activity.
- [ ] The project has a scenario showing money entering the ledger from an external source.
- [ ] The project has a scenario showing the external reference used to reconcile that deposit.

## System Boundary Requirements

The project should make a deliberate choice about what belongs in Go and what belongs in Postgres. The goal is not "Postgres first." The goal is to put each rule where it can be enforced clearly and safely.

- [x] The docs explain that Postgres owns durable ledger storage.
- [x] The docs explain that Postgres owns foreign keys, uniqueness, basic checks, and transaction atomicity.
- [x] The docs explain that Postgres owns final checks that cannot rely on stale application reads.
- [x] The docs explain that Go owns the public command/API shape.
- [x] The docs explain that Go owns request parsing and request-level validation.
- [x] The docs explain that Go owns idempotency behavior from the caller's point of view.
- [x] The docs explain that Go owns error mapping into stable application errors.
- [x] The docs explain that Go owns orchestration for multi-step workflows.
- [x] The docs explain that Go owns retry behavior, worker behavior, and external integration flow.
- [x] The docs explain which invariants are enforced in the database and which rules are enforced in Go.
- [x] The docs explain why each boundary was chosen.
- [ ] The project has at least one ledger operation exposed through Go instead of only direct `psql` calls.
- [ ] The project still proves the underlying database behavior with SQL scenarios.

## Reversal Requirements

- [ ] A posted transaction can be reversed with a new transaction.
- [ ] A reversal creates new entries instead of editing old entries.
- [ ] The original transaction remains visible and unchanged.
- [ ] A transaction cannot be reversed twice.
- [ ] Reversal entries balance to zero.
- [ ] Reversal behavior has a scenario file with expected output.

## Balance And Reconciliation Requirements

- [x] There is a query or view that derives balances from entries.
- [x] There is a scenario comparing derived balances to `ledger_accounts.balance`.
- [ ] Balance comparison should return no mismatches after valid scenarios.
- [ ] The docs explain why stored balances exist if entries are the audit record.
- [ ] The docs explain what happens if stored balances and derived balances disagree.
- [ ] There is a true-up or repair plan, even if it is manual for this project.
- [ ] The project names which transaction states count toward derived balances.
- [ ] External funding or deposit transactions can be tied back to an external reference.
- [ ] The project can identify ledger deposits that do not have a matching external event.
- [ ] The project can identify external events that were not posted to the ledger.

## Database Safety Requirements

- [x] Transfers lock the source account before checking balance.
- [x] Transfers lock the destination account before posting entries.
- [x] Request identity fields are stored on `ledger_transactions` for idempotency checks.
- [x] `idempotency_key` is unique.
- [ ] Amounts must be positive where business rules require positive request amounts.
- [ ] Entry amounts must never be zero.
- [ ] Account currency must match transaction currency.
- [ ] A transaction should not move money across currencies unless an FX flow exists.
- [ ] Transaction `type` should be constrained to known values.
- [ ] Currency codes should be constrained to a known format.
- [ ] The project should document whether it relies on constraints, triggers, permissions, or function-only writes for each invariant.
- [ ] The project should include at least one negative scenario that proves unsafe writes fail.

## Go Service Requirements

The Go layer does not need to be a full production API, but it should be real enough to show how an application would use the ledger safely.

- [ ] Define Go types for ledger commands, results, and application errors.
- [ ] Expose a transfer command through a small HTTP API or CLI.
- [ ] Expose a deposit or funding command through a small HTTP API or CLI.
- [ ] Validate request shape before calling the database.
- [ ] Pass idempotency keys through from the caller to the posting path.
- [ ] Return the original transaction for a repeated idempotent request.
- [ ] Return a clear conflict error for idempotency-key misuse.
- [ ] Return a clear insufficient-funds error.
- [ ] Keep database errors from leaking directly to callers.
- [ ] Use database transactions deliberately from Go where Go owns orchestration.
- [ ] Add tests around the Go boundary for success, retry, conflict, and insufficient funds.
- [ ] Document which ledger rules Go must not try to enforce by stale preflight reads.

## Scenario Requirements

Each important rule should have a small SQL scenario that can be run through `psql`.

- [x] Happy path transfer.
- [ ] Money entering the ledger from an external source.
- [x] Insufficient funds.
- [x] Sequential idempotency retry.
- [x] Concurrent idempotency retry.
- [x] Reusing an idempotency key with different request fields.
- [x] Stored balance versus derived balance comparison.
- [ ] Unbalanced transaction prevention.
- [ ] Direct write protection or documented direct-write risk.
- [ ] Reversal happy path.
- [ ] Double reversal failure.
- [ ] Mutation protection for posted entries.
- [ ] Mutation protection for posted transactions.

## Documentation Requirements

- [ ] Explain the schema in plain language.
- [ ] Explain the lifecycle of a transaction from request to posted entries.
- [ ] Explain the system boundary between Go and Postgres.
- [ ] Explain how money enters the ledger from outside systems.
- [ ] List the invariants the ledger depends on.
- [ ] For each invariant, say where it is enforced:
  - Go request validation
  - Go service orchestration
  - schema constraint
  - foreign key
  - unique index
  - trigger
  - stored function
  - database permissions
  - scenario test only
- [ ] Link to the scenario that proves each implemented behavior.
- [ ] Explain the difference between stored balances and derived balances.
- [ ] Explain why reversals are new transactions instead of edits.
- [ ] Call out which shortcuts are learning shortcuts.
- [ ] Call out which missing features would be required for production.

## Industry Features To Recognize

These do not all need to be implemented in this project, but the final docs should show that they are understood.

- [ ] Multiple ledgers or ledger namespaces.
- [ ] Pending versus posted balances.
- [ ] Holds, authorizations, and settlement.
- [ ] External payment reconciliation.
- [ ] Idempotency across API retries and worker retries.
- [ ] Exactly-once effect built from at-least-once execution.
- [ ] Balance snapshots for faster historical reads.
- [ ] Account status and account closure rules.
- [ ] Negative-balance policy by account type.
- [ ] Multi-currency and FX handling.
- [ ] Decimal or minor-unit strategy by currency.
- [ ] Audit trails for who or what initiated each transaction.
- [ ] Role-based write permissions.
- [ ] Internal service authentication and authorization.
- [ ] Migration strategy that does not drop live ledger tables.
- [ ] Backfills and repair jobs.
- [ ] Operational checks for balance mismatches.
- [ ] Monitoring and alerting for failed posting attempts.
- [ ] Disaster recovery and backup/restore expectations.
- [ ] Data retention and compliance requirements.

## Out Of Scope For Implementation

These are not required to build during this phase.

- Multiple ledgers.
- Pending versus posted balances.
- FX.
- Payment processor integrations.
- Full migration tooling.
- Production auth and permissions.
- Production monitoring.
- Disaster recovery automation.

## Exercise Shortcut

`add_balance` uses an internal `External Funding` account so deposits can satisfy the current `from_account_id` constraint and still create balanced entries.

That account is a learning-project shortcut. The real version should model external money movement more deliberately:

- The outside bank account, card network, or processor is not itself a normal user ledger account.
- The ledger uses an internal cash, settlement, clearing, or receivable account to keep entries balanced.
- The external payment event is stored as a separate reference and reconciled against the ledger transaction.
