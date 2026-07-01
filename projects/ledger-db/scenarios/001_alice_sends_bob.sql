-- local/dev only
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

-- RESULT
-- TRUNCATE TABLE
-- INSERT 0 2
--  transaction_id
-- ----------------
--               1
-- (1 row)

--  id | name  | currency_code | balance
-- ----+-------+---------------+---------
--   1 | Alice | USD           |    1000
--   2 | Bob   | USD           |    1000
-- (2 rows)

--  id |   type   |   idempotency_key    |          created_at
-- ----+----------+----------------------+-------------------------------
--   1 | transfer | alice-sends-bob-1000 | 2026-06-30 20:51:05.943522-05
-- (1 row)

--  id | transaction_id | account_id | amount
-- ----+----------------+------------+--------
--   1 |              1 |          1 |  -1000
--   2 |              1 |          2 |   1000
-- (2 rows)

-- ledger_db=>