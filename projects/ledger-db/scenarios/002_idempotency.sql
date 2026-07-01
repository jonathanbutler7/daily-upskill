-- local/dev only
truncate table ledger_entries, ledger_transactions, ledger_accounts restart identity;

insert into ledger_accounts(name, description, currency_code, balance)
values 
    ('Alice', 'Alice Wallet', 'USD', 2000),
    ('Bob', 'Bob Wallet', 'USD', 0);

select post_transfer(1, 2, 1000, 'same-request') as first_transaction_id;
select post_transfer(1, 2, 1000, 'same-request') as second_transaction_id;

select * from ledger_accounts order by id;
select * from ledger_transactions order by id;
select * from ledger_entries order by id;

-- RESULT
-- TRUNCATE TABLE
-- INSERT 0 2
--  first_transaction_id
-- ----------------------
--                     1
-- (1 row)

--  second_transaction_id
-- -----------------------
--                      1
-- (1 row)

--  id | name  | description  | currency_code | balance |          created_at
-- ----+-------+--------------+---------------+---------+-------------------------------
--   1 | Alice | Alice Wallet | USD           |    1000 | 2026-06-30 20:47:28.494248-05
--   2 | Bob   | Bob Wallet   | USD           |    1000 | 2026-06-30 20:47:28.494248-05
-- (2 rows)

--  id |   type   | idempotency_key |          created_at
-- ----+----------+-----------------+-------------------------------
--   1 | transfer | same-request    | 2026-06-30 20:47:28.499479-05
-- (1 row)

--  id | transaction_id | account_id | amount |          created_at
-- ----+----------------+------------+--------+-------------------------------
--   1 |              1 |          1 |  -1000 | 2026-06-30 20:47:28.499479-05
--   2 |              1 |          2 |   1000 | 2026-06-30 20:47:28.499479-05
-- (2 rows)