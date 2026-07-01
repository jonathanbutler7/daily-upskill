# Concurrency

How does the `create_post_transfer` pg function handle if 2 requests with the same idempotency key are attempted at the same time?

## Problem

In order to document problematic behavior, force a race condition by adding a `perform pg_sleep(5);` to `create_post_transfer`. 

Then attempt the same db transaction in 2 separate psql instsances within a 5 second window.

If you do that, you will get this result:

### Transaction 1
```sql
ledger_db=> begin;
select post_transfer(1, 2, 1000, 'same-concurrent-request');
commit;
BEGIN
 post_transfer
---------------
             1
(1 row)

COMMIT
```

### Transaction 2

```sql
ledger_db=> begin;
select post_transfer(1, 2, 1000, 'same-concurrent-request');
commit;
BEGIN
ERROR:  duplicate key value violates unique constraint "ledger_transactions_idempotency_key_key"
DETAIL:  Key (idempotency_key)=(same-concurrent-request) already exists.
CONTEXT:  SQL statement "insert into ledger_transactions (type, idempotency_key)
    values ('transfer', idempotency_key)
    returning id"
PL/pgSQL function post_transfer(bigint,bigint,bigint,text) line 67 at SQL statement
ROLLBACK
```

### Behavior

Because there is a unique constraint on the `ledger_transactions` table, the db protects itself from getting into state where money was moved twice.

However, in order to handle the error more gracefully and allow for retries, it's important that the idempotency behavior instead returns the original transaction id instead of an error.

## Solution

The solution is to check if the idempotency already exists when creating the ledger_transactions record. If it does, return the transaction id.

### Transaction 1

```sql
ledger_db=> begin;
select post_transfer(1, 2, 1000, 'same-concurrent-request');
commit;
BEGIN
 post_transfer
---------------
             1
(1 row)
```

### Transaction 2

```sql
ledger_db=> begin;
select post_transfer(1, 2, 1000, 'same-concurrent-request');
commit;
BEGIN
 post_transfer
---------------
             1
(1 row)
```