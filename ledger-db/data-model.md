# Data Model

## Entities
- ledger
  - id
  - name
  - description
  - metadata
  - created_at
  - updated_at
- account
  - id
  - name
- ledger_transaction
  - id
  - timestamp
  - amount
  - type
  - status
  - balance
- ledger_entry
  - amount
  - direction: CREDIT | DEBIT
  - ledger_account


## Rules
- posted transaction entries must sum to 0
- posted transaction must have at least 2 entries
- account balance cannot go below 0 unless allowed
- duplicate idempotency key must not post twice
- reversal references original transaction
- original transaction is never edited

## Questions for Postgres
- Which rules can be enforced with constraints?
- Which rules need transaction logic?
- Which rules need isolation/locking?
- Which rules does Go need to handle?

## First API / CLI commands
- create account
- post transfer
- get account balance
- reverse transaction