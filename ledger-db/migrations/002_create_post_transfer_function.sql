drop function if exists post_transfer(bigint, bigint, bigint, text);

create function post_transfer(
    from_account_id bigint,
    to_account_id bigint,
    transfer_amount bigint,
    idempotency_key text
)
returns bigint
language plpgsql
as $$
declare
    existing_transaction_id bigint;
    new_transaction_id bigint;
    from_balance bigint;
    from_currency char(3);
    to_currency char(3);
begin

    -- Manually validate idempotency_key is unique
    -- for better UX. Postgres will prevent duplicates
    -- by returning a UNIQUE constraint error on the
    -- idempotency_key field. But the product requirement
    -- is to return a value rather than an error. Handle
    -- it here rather than application layer because a
    -- request retry should not move money twice.
    select id
        into existing_transaction_id
    from ledger_transactions
    where ledger_transactions.idempotency_key = post_transfer.idempotency_key;

    if existing_transaction_id is not null then
        return existing_transaction_id;
    end if;

    -- Validate amount > 0.
    if transfer_amount <= 0 then
        raise exception 'amount must be greater than zero';
    end if;

    -- Look up the from account.
    select balance, currency_code
    into from_balance, from_currency
    from ledger_accounts
    where id = from_account_id
    for update;

    if not found then 
        raise exception 'from account not found';
    end if;

    -- Look up the to account.
    select currency_code
    into to_currency
    from ledger_accounts
    where id = to_account_id
    for update;

    if not found then
        raise exception 'to account not found';
    end if;
    
    -- Check currencies match.
    if from_currency <> to_currency then
        raise exception 'currency mismatch';
    end if;

    -- Check sufficient balance.
    if from_balance < transfer_amount then
        raise exception 'insufficient funds';
    end if;
    
    -- Insert transaction.
    insert into ledger_transactions (type, idempotency_key)
    values ('transfer', idempotency_key)
    returning id into new_transaction_id;

    -- Insert entries.
    insert into ledger_entries (transaction_id, account_id, amount)
    values
        (new_transaction_id, from_account_id, -transfer_amount),
        (new_transaction_id, to_account_id, transfer_amount);

    -- Update balances.
    update ledger_accounts
    set balance = balance - transfer_amount
    where id = from_account_id;

    update ledger_accounts
    set balance = balance + transfer_amount
    where id = to_account_id;

    return new_transaction_id;
end;
$$;
