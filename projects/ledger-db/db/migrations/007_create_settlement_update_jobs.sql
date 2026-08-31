drop table if exists settlement_update_jobs;

create table settlement_update_jobs (
    id bigserial primary key,
    ledger_entry_id bigint not null unique references ledger_entries(id),
    settlement_account_id bigint not null references ledger_accounts(id),
    amount bigint not null check (amount <> 0),
    status text not null default 'pending' check (status in ('pending', 'applied')),
    attempts integer not null default 0,
    last_error text,
    created_at timestamptz not null default now(),
    processed_at timestamptz
);

create index settlement_update_jobs_pending_idx
    on settlement_update_jobs (id)
    where status = 'pending';
