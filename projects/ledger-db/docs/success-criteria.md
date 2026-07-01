# Success Criteria

The goal of this project is to build a small Postgres-first double-entry ledger.

This is not meant to become a full payments platform. It is a homework project for learning how ledger data should be modeled, what the database can enforce, and where money movement can go wrong.

## Done means

The project is done when it can safely model basic money movement and explain the rules it depends on.

Specifically:

- [x] The schema has accounts, ledger transactions, and ledger entries.
- [x] Money movement happens through Postgres functions, not hand-written ad hoc inserts.
- [x] A valid transfer creates one ledger transaction and two balanced entries.
- [x] Insufficient funds fails without moving money.
- [x] Same idempotency key with the same request returns the original transaction.
- [x] Same idempotency key with a different request fails.
- [x] Concurrent retries with the same request move money once.
- [ ] Reversals create new entries instead of editing old ones.
- [ ] Stored balances can be compared against balances derived from entries.
- [ ] The docs explain what Postgres enforces, what the functions enforce, and what is out of scope.

## Milestones

1. Finish `post_transfer`.
   - [x] Valid transfer works.
   - [x] Insufficient funds fails.
   - [x] Sequential idempotency works.
   - [x] Concurrent idempotency works.
   - [x] Reusing an idempotency key with different request fields fails.
   - [ ] Add scenario output for mismatched idempotency keys.
   - [ ] Add scenario output for insufficient funds.
   - [ ] Add notes for the concurrent idempotency test.

2. Add balance checks.
   - [ ] Write a query or view that derives account balances from entries.
   - [ ] Compare derived balances to `ledger_accounts.balance`.
   - [ ] Add a true-up query that finds mismatches.

3. Prove entry balancing.
   - [ ] Show that each posted ledger transaction should sum to zero.
   - [ ] Decide how this project prevents unbalanced writes.
   - [ ] Document the tradeoff: function-only writes, trigger, or validation query.

4. Add reversals.
   - [ ] Reverse a transaction with a new transaction.
   - [ ] Do not edit the original transaction or entries.
   - [ ] Prevent double reversal.

5. Write the final project note.
   - [ ] Explain the schema.
   - [ ] List the invariants.
   - [ ] Link to the scenarios that prove behavior.
   - [ ] Call out the limits of this design.

## Out of scope

- Go API code
- Production auth and permissions
- Multiple ledgers
- Pending versus posted balances
- FX (Foreign Exchange)
- Payment processor integrations
- Full migration tooling
