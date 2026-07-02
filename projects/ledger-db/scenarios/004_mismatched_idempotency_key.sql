-- local/dev only
truncate table ledger_entries, ledger_transactions, ledger_accounts restart identity;

insert into ledger_accounts(name, description, currency_code, balance)
values
    ('Alice', 'Alice Wallet', 'USD', 5000),
    ('Bob', 'Bob Wallet', 'USD', 0);

select post_transfer(1, 2, 1000, 'same-key') as first_transaction_id;
select post_transfer(1, 2, 2000, 'same-key') as mismatched_request;

-- RESULT
-- TRUNCATE TABLE
-- INSERT 0 2
--  first_transaction_id
-- ----------------------
--                     1
-- (1 row)
--
-- ERROR:  idempotency key reused with different request
-- CONTEXT:  PL/pgSQL function post_transfer(bigint,bigint,bigint,text) line 69 at RAISE
