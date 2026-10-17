alter table outbox_events add column if not exists processed_at timestamptz;

alter table outbox_events drop constraint if exists outbox_events_event_type_check;
alter table outbox_events add constraint outbox_events_event_type_check check (
    event_type in ('ledger.entry.posted', 'ledger.transaction.posted',
                   'ledger.transaction.reversed', 'ledger.settlement_balance.requested')
);
alter table outbox_events add constraint outbox_settlement_source_check check (
    event_type <> 'ledger.settlement_balance.requested' or source_type = 'ledger_entry'
);

-- One balance application per immutable settlement entry.
create unique index outbox_settlement_entry_idx on outbox_events (source_id)
    where event_type = 'ledger.settlement_balance.requested';
create index outbox_settlement_pending_idx on outbox_events (id)
    where event_type = 'ledger.settlement_balance.requested' and processed_at is null;

create or replace function notify_settlement_outbox() returns trigger
language plpgsql as $$
begin
    perform pg_notify('ledger_settlement', '');
    return new;
end;
$$;

create trigger settlement_outbox_inserted after insert on outbox_events
    for each row when (new.event_type = 'ledger.settlement_balance.requested')
    execute function notify_settlement_outbox();
