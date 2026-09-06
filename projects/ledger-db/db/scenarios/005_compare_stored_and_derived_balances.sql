-- local/dev only
truncate table external_transfers, ledger_reversals, ledger_entries, ledger_transactions, ledger_accounts restart identity;

\ir ../migrations/003_seed_system_accounts.sql

insert into ledger_accounts(name, description, currency_code, normal_balance, ledgerable_type, balance)
values
    ('Alice', 'Alice Wallet', 'USD', 'credit', 'internal_account', 0),
    ('Bob', 'Bob Wallet', 'USD', 'credit', 'internal_account', 0);

select deposit_funds(2, 2000, 'ach', 'alice-balance-check-seed-ext', 'alice-balance-check-seed') as deposit_transaction_id;
select post_transfer(2, 3, 1000, 'alice-sends-bob-1000') as first_transfer_transaction_id;
select post_transfer(2, 3, 200, 'alice-sends-bob-200') as second_transfer_transaction_id;

select
    le.account_id,
    sum(
        case
            when le.direction = la.normal_balance then le.amount
            else -le.amount
        end
    ) as derived_balance,
    la.balance as stored_balance
from ledger_entries le
join ledger_accounts la on la.id = le.account_id
group by le.account_id, la.balance, la.normal_balance
order by le.account_id;

-- RESULT
-- TRUNCATE TABLE
-- INSERT 0 1
-- INSERT 0 2
--  deposit_transaction_id
-- ------------------------
--                       1
-- (1 row)
--
--  first_transfer_transaction_id
-- -------------------------------
--                              2
-- (1 row)
--
--  second_transfer_transaction_id
-- --------------------------------
--                               3
-- (1 row)
--
--  account_id | derived_balance | stored_balance
-- ------------+-----------------+----------------
--          1 |            2000 |           2000
--          2 |             800 |            800
--          3 |            1200 |           1200
-- 
