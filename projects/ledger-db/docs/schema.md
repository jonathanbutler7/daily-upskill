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
- currency_code
- balance: number

#### Out of scope
- allow_negative_balance 
- user_id

### Transaction
the business event container
- id PK
- type: transfer | reversal
- idempotency_key
- created_at

### Entry
child record of a transaction. 
the accounting lines inside the transaction event
- id PK
- transaction_id FK -> transaction.id
- account_id FK -> account.id
- amount
- created_at

### Reversal
a transaction initiated internally to correct a ledger
- id PK
- reversing_transaction_id FK -> transaction.id
- original_transaction_id FK -> transaction.id
- reason
