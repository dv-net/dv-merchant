CREATE TABLE IF NOT EXISTS exchange_withdrawal_history (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL,
    exchange_id uuid NOT NULL,
    exchange_order_id VARCHAR(255) NULL,
    address VARCHAR(255) NOT NULL,
    native_amount numeric(30, 15) NULL,
    fiat_amount numeric(30, 15) NULL,
    currency VARCHAR(255) NOT NULL,
    chain VARCHAR(255) NOT NULL,
    status VARCHAR(255) NOT NULL,
    txid VARCHAR(255) NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP
);