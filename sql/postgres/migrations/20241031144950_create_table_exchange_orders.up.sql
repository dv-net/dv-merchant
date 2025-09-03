CREATE TABLE IF NOT EXISTS exchange_orders (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    exchange_id uuid NOT NULL,
    exchange_order_id varchar(255) NULL,
    client_order_id varchar(255) NULL,
    symbol VARCHAR(255) NOT NULL,
    side VARCHAR(255) NOT NULL,
    amount DECIMAL NOT NULL,
    order_created_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);