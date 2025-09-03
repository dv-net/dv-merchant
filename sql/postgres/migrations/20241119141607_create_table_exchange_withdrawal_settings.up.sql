CREATE TABLE exchange_withdrawal_settings (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL,
    exchange_id uuid NOT NULL,
    currency VARCHAR(255) NOT NULL,
    chain VARCHAR(255) NOT NULL,
    address VARCHAR(255) NOT NULL,
    min_amount numeric(30, 15) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP
);