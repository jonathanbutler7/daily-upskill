drop table if exists outbox_events;

create table outbox_events (
    id bigserial primary key,
    occurred_at timestamptz,
    created_at timestamptz not null default now(),
    -- Describes where the outbox event originated from.
    -- Could be `ledger_transaction`, `ledger_entry`, etc.
    source_type text not null check (
        source_type in ('ledger_entry', 'ledger_transaction', 'ledger_reversal')
    ),
    -- The identifier for the entity that this outbox event
    -- is related to.
    source_id bigint not null,
    -- Type of event being entered into the outbox.
    event_type text not null check (
        event_type in (
        'ledger.entry.posted',
        'ledger.transaction.posted',
        'ledger.transaction.reversed'
        )
    ),
    attempt_count bigint not null default 0,
    payload jsonb not null
);

create index outbox_events_source_idx
    on outbox_events (source_type, source_id);