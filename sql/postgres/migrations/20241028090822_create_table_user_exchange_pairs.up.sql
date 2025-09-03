CREATE TABLE IF NOT EXISTS user_exchange_pairs (
    id uuid PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    exchange_id uuid NOT NULL,
    user_id uuid NOT NULL,
    currency_from varchar(255) NOT NULL,
    currency_to varchar(255) NOT NULL,
    symbol varchar(255) NOT NULL,
    type varchar(255) NOT NULL
);