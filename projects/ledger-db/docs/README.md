# Ledger database

Ledger database uses double entry bookkeeping to provide an immutable, fully auditable system of record that tracks transfers of funds.

## Transaction lifecycles

### Internal transfers

transfer of funds from user account -> user account

1. ledger validates args on request
2. ledger begins db transaction
3. ledger applies row locks to and from accounts to avoid write conflicts
4. ledger checks to make sure the transaction is idempotent
   1. same request returns same result
   2. different request returns error
5. ledger makes sure that the balance is available to transfer
6. ledger inserts entries for from account and to account
7. ledger verifies transaction balances
8. ledger adjusts account balances
9. ledger commits or rolls back db transaction

### Withdrawals

transfer of funds from user account -> Cash settlement

1. ledger validates args on request
2. ledger begins db transaction
3. ledger locks to account and cash settlement account to avoid write conflicts
4. ledger guarantees the transaction is idempotent
   1. same request returns same result 
   2. conflicting request returns error
5. ledger inserts transaction
   1. ensures idempotency
6. ledger creates entries for to account and cash settlement account
7. ledger adjusts account balances
8. ledger creates external transfer records
9. ledger commits or rolls back db transaction

### Deposits

transfer of funds from Cash settlement -> user account

Currently deposits use the same function as withdrawals, so the steps are the same as above.

### Reversals

JTBD

### Error handling and failed transfers

Transfers that contain invalid requests or fail at the db layer return typed errors to inform the caller of the result, and allow the caller to handle retry strategies.

## System guarantees

### Ledger operations

1. `PostExternalTransfer` 
   1. handles external depositing and withdrawing of funds
2. `PostTransfer`
   1. handles internal transfers between 2 accounts
3. `Reversal`
   1. JTBD

### Atomicity

Each ledger operation is wrapped in a single database transaction, to ensure data correctness and prevent write conflicts. If any step of the operation fails, it will be rolled back as a single transaction, and an error will be returned.

### Invariants

### Concurrency

### Immutability

The following tables are immutable so that the ledger db is fully auditable at any state. 

- `ledger_entries`
- `ledger_transactions`

`update` and `delete` are prevented at the postgres level for each row.

The implication is that the tables are append-only.

### Data Correctness

At an interval, ledger db will run a job that verifies all entries transaction pair balances to zero and compare3s stored account balances to balances derived from entries.

Example: JBTD

### Serialization at scale

The cash settlement account, for example, could become hot because it is a single account that all external transfers will need to get a row lock on.

Strategies as the project scales

- Create multiple system accounts to reduce lock contention, so that every external transfer does not touch the same row
- Spread out high volume writes by using per-rail clearing accounts to handle the initial money movement, then reconcile at an interval into the smaller settlement accounts

### Retries

The ledger db returns details error messages, and allows the caller to define a retry strategy.

### Currencies

Currently ledger db only supports transfers, withdrawals, and deposits with the same currency.

A future state of this project will add currency conversion.

### Balance types

- pending
- settled
- available

### Transaction types

1. transfer
2. deposit
3. withdrawal
4. reversal

### Data model

Ledger db is comprised of the following tables

- ledger_accounts
- ledger_transactions
- ledger_entries
- ledger_reversals
- external_transfers

### System accounts

A system cash settlement account (in USD) maintains double entry bookkeeping for external transfers.

### Auditability and immutability

Since the tables in the ledger db are immutable, any single point in the ledger db's history can be audited and reconstructed to validate correctness