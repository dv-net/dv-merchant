CREATE TABLE IF NOT EXISTS update_balance_queue
(
    id                          uuid primary key            not null default gen_random_uuid(),
    currency_id                 varchar(255)                not null
        constraint wallet_addresses_currency_id_foreign
            references currencies,
    address                     varchar(255)                not null,
    native_token_balance_update bool                        not null default false,
    created_at                  timestamp without time zone not null default now(),
    updated_at                  timestamp without time zone          default null
);