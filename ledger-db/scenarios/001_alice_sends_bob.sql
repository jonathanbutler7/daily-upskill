truncate table ledger_entries, ledger_transactions, ledger_accounts restart identity;

insert into ledger_accounts (name, description, currency_code, balance)
values
    ('Alice', 'Alice wallet', 'USD', 2000),
    ('Bob', 'Bob wallet', 'USD', 0);

select post_transfer(1, 2, 1000, 'alice-sends-bob-1000') as transaction_id;

select id, name, currency_code, balance
from ledger_accounts
order by id;

select id, type, idempotency_key, created_at
from ledger_transactions
order by id;

select id, transaction_id, account_id, amount
from ledger_entries
order by id;
