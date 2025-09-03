CREATE TABLE IF NOT EXISTS withdrawal_from_processing_wallets
(
    id           uuid                                 default gen_random_uuid() primary key,
    currency_id  varchar                     not null,
    store_id     uuid                        not null,
    transfer_id  uuid                                 default null,
    address_from varchar                     not null,
    address_to   varchar                     not null,
    amount       numeric(90, 50)             not null,
    amount_usd   numeric(28, 8)              not null,
    created_at   timestamp without time zone not null default now(),
    updated_at   timestamp without time zone          default null
);