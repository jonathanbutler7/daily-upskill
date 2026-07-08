-- Deprecated: the Go path now uses cmd.PostTransfer -> ledgerstore.PostTransfer
-- instead of calling this Postgres function directly. Keep this only as a
-- reference while the migration is in progress.

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

    -- Look up and lock the from account.
    select balance, currency_code
    into from_balance, from_currency
    from ledger_accounts
    where id = from_account_id
    for update;

    if not found then
        raise exception 'from account not found';
    end if;

    -- Look up and lock the to account.
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

    -- If this exact request already posted, return its transaction id.
    select id
    into existing_transaction_id
    from ledger_transactions lt
    where lt.idempotency_key = post_transfer.idempotency_key
        and lt.from_account_id = post_transfer.from_account_id
        and lt.to_account_id = post_transfer.to_account_id
        and lt.amount = post_transfer.transfer_amount
        and lt.currency_code = from_currency;

    if existing_transaction_id is not null then
        return existing_transaction_id;
    end if;

    -- If the key exists for a different request, reject it.
    select id
    into existing_transaction_id
    from ledger_transactions lt
    where lt.idempotency_key = post_transfer.idempotency_key
        and not (
            lt.from_account_id = post_transfer.from_account_id
            and lt.to_account_id = post_transfer.to_account_id
            and lt.amount = post_transfer.transfer_amount
            and lt.currency_code = from_currency
        );

    if existing_transaction_id is not null then
        raise exception 'idempotency key reused with different request';
    end if;
    
    -- pg sleep allows us to test concurrent transaction and make sure
    -- race conditions are handled correctly by idempotency
    -- perform pg_sleep(5);

    -- Check balance
    if from_balance < transfer_amount then
        raise exception 'insufficient funds';
    end if;

    -- Insert transaction.
    begin
        insert into ledger_transactions (
            type, 
            idempotency_key, 
            from_account_id, 
            to_account_id, 
            amount, 
            currency_code
        )
        values (
            'transfer',
            idempotency_key, 
            from_account_id,
            to_account_id,
            transfer_amount,
            from_currency
        )
        returning id into new_transaction_id;
    exception
        when unique_violation then
            select id
            into existing_transaction_id
            from ledger_transactions lt
            where lt.idempotency_key = post_transfer.idempotency_key
                and lt.from_account_id = post_transfer.from_account_id
                and lt.to_account_id = post_transfer.to_account_id
                and lt.amount = post_transfer.transfer_amount
                and lt.currency_code = from_currency;

            if existing_transaction_id is not null then
                return existing_transaction_id;
            end if;

            select id
            into existing_transaction_id
            from ledger_transactions lt
            where lt.idempotency_key = post_transfer.idempotency_key;

            if existing_transaction_id is not null then
                raise exception 'idempotency key reused with different request';
            end if;

            raise;
    end;

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
