# Schema

## Nouns

- Account
- Transaction
- Entry
- Reversal

## Fields

### Account
the container for sending or receiving money
- id PK
- name: string
- description: string
- balance: number (cents)
- normal_balance: debit | credit
- user_id: no user entity here, but this calls out the gap

### Transaction
the business event container
- id PK
- type: transfer | reversal
- status: posted | fail | etc.
- idempotency_key

### Entry
child record of a transaction. 
the accounting lines inside the transaction event
- id PK
- transaction_id FK -> transaction.id
- account_id FK -> account.id
- amount

### Reversal
a transaction initiated internally to correct a ledger
- id PK
- reversing_transaction_id FK -> transaction.id
- original_transaction_id FK -> transaction.id
- reason
