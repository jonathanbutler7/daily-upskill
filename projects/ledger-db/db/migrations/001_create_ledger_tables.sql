drop table if exists external_transfers;
drop table if exists ledger_entries;
drop table if exists ledger_transactions;
drop table if exists ledger_accounts;
drop table if exists ledger_reversals;

create table ledger_accounts (
    id bigserial primary key,
    name text not null,
    description text not null,
    currency_code char(3) not null,
    balance bigint not null default 0,
    -- allow_negative_balance boolean not null default false,
    created_at timestamptz not null default now()
);

create table ledger_transactions (
    id bigserial primary key,
    type text not null check(type in ('transfer', 'deposit', 'withdrawal', 'reversal')),
    idempotency_key text not null unique,
    created_at timestamptz not null default now(),
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
