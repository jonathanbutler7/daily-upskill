drop trigger if exists ledger_entries_no_update on ledger_entries;
drop function if exists prevent_ledger_entry_mutation();
drop trigger if exists ledger_transactions_no_update on ledger_transactions;
drop function if exists prevent_ledger_transaction_mutation();

create function prevent_ledger_entry_mutation()
returns trigger
language plpgsql
as $$
begin
    raise exception 'posted ledger entries are immutable';
end;
$$;

create trigger ledger_entries_no_update
before update or delete on ledger_entries
for each row
execute function prevent_ledger_entry_mutation();

create function prevent_ledger_transaction_mutation()
returns trigger
language plpgsql
as $$
begin
    raise exception 'posted ledger transactions are immutable';
end;
$$;

create trigger ledger_transactions_no_update
before update or delete on ledger_transactions
for each row
execute function prevent_ledger_transaction_mutation()