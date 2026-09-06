drop function if exists deposit_funds(bigint, bigint, text, text, text);
drop function if exists post_transfer(bigint, bigint, bigint, text);

drop table if exists external_transfers;
drop table if exists ledger_reversals;
drop table if exists settlement_update_jobs;
drop table if exists ledger_entries;
drop table if exists ledger_transactions;
drop table if exists ledger_accounts;

create table ledger_accounts (
    id bigserial primary key,
    name text not null,
    description text not null,
    currency_code char(3) not null,
    normal_balance text check(normal_balance in ('credit', 'debit')),
    ledgerable_type text check(ledgerable_type in ('external_account', 'internal_account')),
    lock_version bigint not null default 0,
    balance bigint not null default 0,
    created_at timestamptz not null default now()
);

create table ledger_transactions (
    id bigserial primary key,
    type text not null check(type in ('transfer', 'deposit', 'withdrawal', 'reversal')),
    status text not null default 'posted' check(status in ('pending', 'posted', 'archived')),
    idempotency_key text not null unique,
    created_at timestamptz not null default now(),
    posted_at timestamptz not null default now(),
    archived_at timestamptz,
    from_account_id bigint not null references ledger_accounts(id),
    to_account_id bigint not null references ledger_accounts(id),
    amount bigint not null check (amount > 0),
    currency_code char(3) not null
);

create table ledger_entries (
    id bigserial primary key,
    transaction_id bigint not null references ledger_transactions(id),
    account_id bigint not null references ledger_accounts(id),
    amount bigint not null check (amount <> 0),
    direction text not null check(direction in ('credit', 'debit')),
    archived boolean not null default false,
    archived_at timestamptz,
    created_at timestamptz not null default now()
);

create table ledger_reversals (
    id bigserial primary key,
    original_transaction_id bigint not null references ledger_transactions(id),
    reversal_transaction_id bigint not null unique references ledger_transactions(id),
    reason text not null,
    created_at timestamptz not null default now(),

    unique (original_transaction_id)
);
