-- local/dev only
truncate table external_transfers, ledger_entries, ledger_transactions, ledger_accounts restart identity;

\ir ../migrations/003_seed_system_accounts.sql

insert into ledger_accounts(name, description, currency_code, balance)
values
    ('Alice', 'Alice Wallet', 'USD', 0),
    ('Bob', 'Bob Wallet', 'USD', 0);

select add_balance(2, 5000, 'ach', 'alice-mismatch-seed-ext', 'alice-mismatch-seed') as deposit_transaction_id;
select post_transfer(2, 3, 1000, 'same-key') as first_transaction_id;
select post_transfer(2, 3, 2000, 'same-key') as mismatched_request;

-- RESULT
-- TRUNCATE TABLE
-- INSERT 0 1
-- INSERT 0 2
--  deposit_transaction_id
-- ------------------------
--                       1
-- (1 row)
--
--  first_transaction_id
-- ----------------------
--                     2
-- (1 row)
--
-- ERROR:  idempotency key reused with different request
-- CONTEXT:  PL/pgSQL function post_transfer(bigint,bigint,bigint,text) line 64 at RAISE
