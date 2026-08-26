CREATE TABLE refund_requests
(
    id                     uuid primary key default gen_random_uuid(),
    blocked_transaction_id uuid unique  not null references blocked_transactions (id),
    wallet_id              uuid         not null references wallets (id),
    store_id               uuid         not null references stores (id),
    transfer_id            uuid,
    destination_address    varchar(255) not null check (destination_address != ''),
    status         varchar(255)  not null check (status != ''),
    email          varchar(255)  not null check (email != ''),
    reviewed_at    timestamp,
    created_at     timestamp     not null,
    updated_at     timestamp
);