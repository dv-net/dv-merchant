CREATE TABLE IF NOT EXISTS exchange_chains (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    slug VARCHAR(255) NOT NULL,
    currency_id VARCHAR(255) NOT NULL,
    ticker VARCHAR(255) NOT NULL,
    chain VARCHAR(255) NOT NULL
);