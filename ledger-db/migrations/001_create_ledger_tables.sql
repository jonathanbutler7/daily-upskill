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
    -- type will be transfer | reversal
    type text not null,
    idempotency_key text not null unique,
    created_at timestamptz not null default now()
);

create table ledger_entries (
    id bigserial primary key,
    transaction_id bigint not null references ledger_transactions(id),
    account_id bigint not null references ledger_accounts(id),
    amount bigint not null check (amount <> 0),
    created_at timestamptz not null default now()
);