insert into ledger_accounts (name, description, currency_code, normal_balance, ledgerable_type, balance)
select
    'Cash Settlement',
    'Internal account used to balance settled external money movement',
    'USD',
    'debit',
    'external_account',
    0
where not exists (
    select 1
    from ledger_accounts
    where name = 'Cash Settlement'
        and currency_code = 'USD'
);
