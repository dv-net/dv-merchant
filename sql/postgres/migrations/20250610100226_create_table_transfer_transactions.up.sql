CREATE TABLE IF NOT EXISTS transfer_transactions
(
    id                  uuid                     not null primary key default gen_random_uuid(),
    transfer_id         uuid                     not null
        constraint fk_transfers_uuid references transfers ON DELETE CASCADE,
    tx_hash             varchar(100)             not null,
    bandwidth_amount    numeric(150, 50)         not null             default 0,
    energy_amount       numeric(150, 50)         not null             default 0,
    native_token_amount numeric(150, 50)         not null             default 0,
    native_token_fee    numeric(150, 50)         not null             default 0,
    tx_type             varchar(255)             not null check (tx_type != ''),
    status              varchar(255)             not null check (status != ''),
    step                varchar(255)             not null check (step != ''),
    created_at          timestamp with time zone not null             default (timezone('utc', now())),
    updated_at          timestamp with time zone not null             default (timezone('utc', now())),

    UNIQUE (transfer_id, tx_hash, tx_type)
);

-- Ordering to optimize queries in internal/escanner/scanner.go (WHERE tx_hash = ? AND tx_type = ?)
CREATE INDEX IF NOT EXISTS tx_hash_tx_type_idx on transfer_transactions(tx_hash, tx_type)