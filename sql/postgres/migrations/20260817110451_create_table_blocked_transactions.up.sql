CREATE TABLE blocked_transactions
(
    id             uuid primary key       default gen_random_uuid(),
    user_id        uuid          not null references users (id),
    store_id       uuid          not null references stores (id),
    transaction_id uuid unique   not null,
    aml_check_id   uuid          not null references aml_checks (id),
    wallet_id      uuid          not null references wallets (id),
    risk_level     varchar(255)  not null check (risk_level != ''),
    score          numeric(6, 3) not null default 0 check (score >= 0),
    created_at     timestamp     not null,
    updated_at     timestamp
);