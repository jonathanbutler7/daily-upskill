drop table if exists external_transfers;

create table external_transfers (
    id bigserial primary key,
    direction text not null check (direction in ('deposit', 'withdrawal')),
    rail text not null check (rail in ('ach', 'instant')),
    status text not null check (status in ('pending', 'posted', 'failed', 'canceled')),
    external_reference text not null unique,
    user_account_id bigint not null references ledger_accounts(id),
    ledger_transaction_id bigint references ledger_transactions(id),
    amount bigint not null check (amount > 0),
    currency_code char(3) not null,
    created_at timestamptz not null default now(),
    completed_at timestamptz
);