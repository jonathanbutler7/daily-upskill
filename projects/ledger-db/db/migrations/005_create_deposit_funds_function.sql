-- Deprecated: the Go path now uses cmd.DepositFunds -> ledgerstore.AddDeposit
-- instead of calling this Postgres function directly. Keep this only as a
-- reference while the migration is in progress.

drop function if exists deposit_funds(bigint, bigint, text, text, text);

create function deposit_funds(
    to_account_id bigint,
    transfer_amount bigint,
    rail text,
    external_reference text,
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
    -- Validate the amount is greater than zero.
    if transfer_amount <= 0 then
        raise exception 'amount must be greater than zero';
    end if;
    
    -- Validate the external_reference value.
    if external_reference is null or btrim(external_reference) = '' then
        raise exception 'external reference must not be empty';
    end if;

    -- Validate the to_account_id exists and lock row. DONE
    select currency_code
    into to_currency
    from ledger_accounts
    where id = to_account_id
    for update;

    if not found then
        raise exception 'to account not found';
    end if;

    -- Locate the internal settlement account. DONE
    select id
    into funding_account_id
    from ledger_accounts
    where name = 'Cash Settlement'
        and currency_code = to_currency
    for update;

    if not found then
        raise exception 'Cash Settlement account not found';
    end if;

    -- Check same idempotency request.
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
    
    -- Check same idempotency conflict. DONE
    select id
    into existing_transaction_id
    from ledger_transactions lt
    where lt.idempotency_key = add_balance.idempotency_key;

    if existing_transaction_id is not null then
        raise exception 'idempotency key reused with different request';
    end if;

    -- Insert ledger transaction. DONE
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
    
    -- Insert ledger entries.
    insert into ledger_entries (transaction_id, account_id, amount)
    values
        (new_transaction_id, funding_account_id, -transfer_amount),
        (new_transaction_id, to_account_id, transfer_amount);
    
    -- Update balances.
    update ledger_accounts
    set balance = balance - transfer_amount
    where id = funding_account_id;

    update ledger_accounts
    set balance = balance + transfer_amount
    where id = to_account_id;

    -- Insert row into external_transfers table.
    insert into external_transfers (
        direction, 
        rail, 
        status, 
        external_reference, 
        user_account_id, 
        ledger_transaction_id, 
        amount,
        currency_code,
        completed_at
    )
    values (
        'deposit', 
        rail, 
        'posted',
        external_reference, 
        to_account_id, 
        new_transaction_id, 
        transfer_amount,
        to_currency,
        now()
    );
    
    return new_transaction_id;
end;
$$;
