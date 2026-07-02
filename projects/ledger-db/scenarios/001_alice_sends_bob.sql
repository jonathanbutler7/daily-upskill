-- local/dev only
truncate table ledger_entries, ledger_transactions, ledger_accounts restart identity;

insert into ledger_accounts (name, description, currency_code, balance)
values
    ('Alice', 'Alice wallet', 'USD', 0),
    ('Bob', 'Bob wallet', 'USD', 0);

select add_balance(1, 2000, 'seed-alice-2000') as deposit_transaction_id;
select post_transfer(1, 2, 1000, 'alice-sends-bob-1000') as transfer_transaction_id;

select id, name, currency_code, balance
from ledger_accounts
order by id;

select id, type, idempotency_key, from_account_id, to_account_id, amount
from ledger_transactions
order by id;

select id, transaction_id, account_id, amount
from ledger_entries
order by id;

select
    account_id,
    sum(amount) as derived_balance
from ledger_entries
group by account_id
order by account_id;

-- RESULT
-- TRUNCATE TABLE
-- INSERT 0 2
--  deposit_transaction_id
-- ------------------------
--                       1
-- (1 row)
--
--  transfer_transaction_id
-- -------------------------
--                        2
-- (1 row)

--  id |       name       | currency_code | balance
-- ----+------------------+---------------+---------
--   1 | Alice            | USD           |    1000
--   2 | Bob              | USD           |    1000
--   3 | External Funding | USD           |   -2000
-- (3 rows)

--  id |   type   |   idempotency_key    | from_account_id | to_account_id | amount
-- ----+----------+----------------------+-----------------+---------------+--------
--   1 | deposit  | seed-alice-2000      |               3 |             1 |   2000
--   2 | transfer | alice-sends-bob-1000 |               1 |             2 |   1000
-- (2 rows)

--  id | transaction_id | account_id | amount
-- ----+----------------+------------+--------
--   1 |              1 |          3 |  -2000
--   2 |              1 |          1 |   2000
--   3 |              2 |          1 |  -1000
--   4 |              2 |          2 |   1000
-- (4 rows)
--
--  account_id | derived_balance
-- ------------+-----------------
--           1 |            1000
--           2 |            1000
--           3 |           -2000
-- (3 rows)

-- ledger_db=>
