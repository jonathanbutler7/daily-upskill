-- local/dev only
truncate table external_transfers, ledger_reversals, ledger_entries, ledger_transactions, ledger_accounts restart identity;

\ir ../migrations/003_seed_system_accounts.sql

insert into ledger_accounts(name, description, currency_code, balance)
values 
    ('Alice', 'Alice Wallet', 'USD', 0),
    ('Bob', 'Bob Wallet', 'USD', 0);

select add_balance(2, 2000, 'ach', 'alice-idempotency-seed-ext', 'alice-idempotency-seed') as deposit_transaction_id;
select post_transfer(2, 3, 1000, 'same-request') as first_transaction_id;
select post_transfer(2, 3, 1000, 'same-request') as second_transaction_id;

select * from ledger_accounts order by id;
select * from ledger_transactions order by id;
select * from ledger_entries order by id;

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

--  second_transaction_id
-- -----------------------
--                      2
-- (1 row)

--  id | name  | description  | currency_code | balance |          created_at
-- ----+-------+--------------+---------------+---------+-------------------------------
--   1 | Cash Settlement | Internal account used to balance settled external money movement | USD | -2000 | ...
--   2 | Alice           | Alice Wallet                                                     | USD |  1000 | ...
--   3 | Bob             | Bob Wallet                                                       | USD |  1000 | ...
-- (3 rows)

--  id |   type   |    idempotency_key     |          created_at           | from_account_id | to_account_id | amount | currency_code
-- ----+----------+------------------------+-------------------------------+-----------------+---------------+--------+---------------
--   1 | deposit  | alice-idempotency-seed | 2026-07-03 21:55:58.838453-05 |               1 |             2 |   2000 | USD
--   2 | transfer | same-request           | 2026-07-03 21:55:58.843556-05 |               2 |             3 |   1000 | USD
-- (2 rows)

--  id | transaction_id | account_id | amount |          created_at
-- ----+----------------+------------+--------+-------------------------------
--   1 |              1 |          1 |  -2000 | 2026-07-03 21:55:58.838453-05
--   2 |              1 |          2 |   2000 | 2026-07-03 21:55:58.838453-05
--   3 |              2 |          2 |  -1000 | 2026-07-03 21:55:58.843556-05
--   4 |              2 |          3 |   1000 | 2026-07-03 21:55:58.843556-05
-- (4 rows)
