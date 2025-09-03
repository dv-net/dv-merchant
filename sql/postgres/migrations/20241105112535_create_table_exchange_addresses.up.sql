CREATE TABLE IF NOT EXISTS exchange_addresses (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    exchange_key_id uuid NOT NULL,
    address varchar(255) NOT NULL,
    chain varchar(255) NOT NULL,
    currency varchar(255) NOT NULL,
    address_type varchar(100) NOT NULL,
    user_id uuid NOT NULL,
    create_type varchar(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);