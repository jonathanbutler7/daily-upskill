drop function if exists add_balance(bigint, bigint, text);

create function add_balance(
    to_account_id bigint,
    transfer_amount bigint,
    idempotency_key text
)
returns bigint
language plpgsql
as $$
declare
    funding_account_id bigint;
    existing_transaction_id bigint;
    new_transaction_id bigint;
    to_currency char(3);
begin
    if transfer_amount <= 0 then
        raise exception 'amount must be greater than zero';
    end if;

    select currency_code
    into to_currency
    from ledger_accounts
    where id = to_account_id
    for update;

    if not found then
        raise exception 'to account not found';
    end if;

    -- This is a learning-project shortcut. The funding account represents
    -- money entering the ledger so deposit entries can still balance to zero.
    select id
    into funding_account_id
    from ledger_accounts
    where name = 'External Funding'
        and currency_code = to_currency
    for update;

    if funding_account_id is null then
        insert into ledger_accounts (name, description, currency_code, balance)
        values (
            'External Funding',
            'Hack for this exercise: source account for deposits and opening balances',
            to_currency,
            0
        )
        returning id into funding_account_id;
    end if;

    select id
    into existing_transaction_id
    from ledger_transactions lt
    where lt.idempotency_key = add_balance.idempotency_key
        and lt.type = 'deposit'
        and lt.from_account_id = funding_account_id
        and lt.to_account_id = add_balance.to_account_id
        and lt.amount = add_balance.transfer_amount
        and lt.currency_code = to_currency;

    if existing_transaction_id is not null then
        return existing_transaction_id;
    end if;

    select id
    into existing_transaction_id
    from ledger_transactions lt
    where lt.idempotency_key = add_balance.idempotency_key;

    if existing_transaction_id is not null then
        raise exception 'idempotency key reused with different request';
    end if;

    insert into ledger_transactions (
        type,
        idempotency_key,
        from_account_id,
        to_account_id,
        amount,
        currency_code
    )
    values (
        'deposit',
        idempotency_key,
        funding_account_id,
        to_account_id,
        transfer_amount,
        to_currency
    )
    returning id into new_transaction_id;

    insert into ledger_entries (transaction_id, account_id, amount)
    values
        (new_transaction_id, funding_account_id, -transfer_amount),
        (new_transaction_id, to_account_id, transfer_amount);

    update ledger_accounts
    set balance = balance - transfer_amount
    where id = funding_account_id;

    update ledger_accounts
    set balance = balance + transfer_amount
    where id = to_account_id;

    return new_transaction_id;
end;
$$;
