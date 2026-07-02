select
    le.account_id,
    sum(le.amount) as derived_balance,
    la.balance as stored_balance
from ledger_entries le
join ledger_accounts la on la.id = le.account_id
group by le.account_id, la.balance
order by le.account_id;

-- RESULT
-- 
-- account_id | derived_balance | stored_balance
-- ------------+-----------------+----------------
--          1 |             800 |            800
--          2 |            1200 |           1200
--          3 |           -2000 |          -2000
-- 
-- CONTEXT
-- ledger_db=> select * from ledger_accounts;
--  id |       name       |                               description                                | currency_code | balance |          created_at
-- ----+------------------+--------------------------------------------------------------------------+---------------+---------+-------------------------------
--   3 | External Funding | Hack for this exercise: source account for deposits and opening balances | USD           |   -2000 | 2026-07-01 22:56:36.881553-05
--   1 | Alice            | Alice wallet                                                             | USD           |     800 | 2026-07-01 22:56:36.880108-05
--   2 | Bob              | Bob wallet                                                               | USD           |    1200 | 2026-07-01 22:56:36.880108-05
--
-- ledger_db=> select * from ledger_entries;
--  id | transaction_id | account_id | amount |          created_at
-- ----+----------------+------------+--------+-------------------------------
--   1 |              1 |          3 |  -2000 | 2026-07-01 22:56:36.881553-05
--   2 |              1 |          1 |   2000 | 2026-07-01 22:56:36.881553-05
--   3 |              2 |          1 |  -1000 | 2026-07-01 22:56:36.885675-05
--   4 |              2 |          2 |   1000 | 2026-07-01 22:56:36.885675-05
--   5 |              3 |          1 |   -200 | 2026-07-01 22:58:43.452567-05
--   6 |              3 |          2 |    200 | 2026-07-01 22:58:43.452567-05