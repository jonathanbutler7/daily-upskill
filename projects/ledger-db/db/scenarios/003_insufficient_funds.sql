-- local/dev only
truncate table external_transfers, ledger_reversals, ledger_entries, ledger_transactions, ledger_accounts restart identity;

\ir ../migrations/003_seed_system_accounts.sql

insert into ledger_accounts(name, description, currency_code, balance)
values 
    ('Alice', 'Alice Wallet', 'USD', 0),
    ('Bob', 'Bob Wallet', 'USD', 0);

select add_balance(2, 2000, 'ach', 'alice-insufficient-seed-ext', 'alice-insufficient-seed') as deposit_transaction_id;
select post_transfer(2, 3, 3000, 'same-key');

-- RESULT
-- TRUNCATE TABLE
-- INSERT 0 1
-- INSERT 0 2
--  deposit_transaction_id
-- ------------------------
--                       1
-- (1 row)
--
-- ERROR:  insufficient funds
-- CONTEXT:  PL/pgSQL function post_transfer(bigint,bigint,bigint,text) line 73 at RAISE
