drop trigger if exists ledger_entries_no_update on ledger_entries;
drop trigger if exists ledger_entries_archive_update on ledger_entries;
drop trigger if exists ledger_entries_no_delete on ledger_entries;
drop trigger if exists ledger_entries_immutable_fields on ledger_entries;
drop trigger if exists ledger_transactions_no_update on ledger_transactions;

drop function if exists prevent_entry_archive_mutation();
drop function if exists prevent_entry_immutable_mutation();
drop function if exists prevent_entry_delete();
drop function if exists prevent_ledger_entry_mutation();
drop function if exists prevent_ledger_transaction_mutation();

create function prevent_entry_archive_mutation()
returns trigger
language plpgsql
as $$
begin
    if old.archived then
        raise exception 'archived ledger entries are immutable';
    end if;

    if not new.archived then
        if new.archived_at is distinct from old.archived_at then
            raise exception 'archived_at can only be set when archiving an entry';
        end if;

        return new;
    end if;

    new.archived_at := coalesce(new.archived_at, now());
    return new;
end;
$$;

create trigger ledger_entries_archive_update
before update of archived, archived_at on ledger_entries
for each row
execute function prevent_entry_archive_mutation();

create function prevent_entry_immutable_mutation()
returns trigger
language plpgsql
as $$
begin
    if tg_op = 'DELETE' then
        raise exception 'ledger entries cannot be deleted';
    end if;

    raise exception 'only archived and archived_at may be updated on ledger entries';
end;
$$;

create trigger ledger_entries_no_delete
before delete on ledger_entries
for each row
execute function prevent_entry_immutable_mutation();

create trigger ledger_entries_immutable_fields
before update of id, transaction_id, account_id, amount, created_at
on ledger_entries
for each row
execute function prevent_entry_immutable_mutation();

create function prevent_ledger_transaction_mutation()
returns trigger
language plpgsql
as $$
begin
    if tg_op = 'DELETE'
        or old.status in ('posted', 'archived') then
        raise exception 'terminal ledger transactions are immutable';
    end if;

    return new;
end;
$$;

create trigger ledger_transactions_no_update
before update or delete on ledger_transactions
for each row
execute function prevent_ledger_transaction_mutation();
