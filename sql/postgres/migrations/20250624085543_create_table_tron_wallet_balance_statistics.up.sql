CREATE TABLE IF NOT EXISTS tron_wallet_balance_statistics
(
    id                  uuid PRIMARY KEY                  default gen_random_uuid(),
    processing_owner_id uuid                     NOT NULL,
    address             varchar(150)             not null check ( address != '' ),
    staked_bandwidth    numeric(150, 50)         not null default 0,
    staked_energy       numeric(150, 50)         not null default 0,
    delegated_energy    numeric(150, 50)         not null default 0,
    delegated_bandwidth numeric(150, 50)         not null default 0,
    available_bandwidth numeric(150, 50)         not null default 0,
    available_energy    numeric(150, 50)         not null default 0,
    created_at          timestamp with time zone not null default (timezone('utc', now())),
    UNIQUE (processing_owner_id, address, created_at)
);

CREATE INDEX IF NOT EXISTS tron_wallet_balance_stats_owner_day_idx
    ON tron_wallet_balance_statistics (processing_owner_id, created_at);