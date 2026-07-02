-- local/dev only
truncate table ledger_entries, ledger_transactions, ledger_accounts restart identity;

insert into ledger_accounts(name, description, currency_code, balance)
values 
    ('Alice', 'Alice Wallet', 'USD', 2000),
    ('Bob', 'Bob Wallet', 'USD', 0);

select post_transfer(1, 2, 3000, 'same-key');

-- RESULT
-- ERROR:  insufficient funds
-- CONTEXT:  PL/pgSQL function post_transfer(bigint,bigint,bigint,text) line 78 at RAISE
