create table if not exists currencies
(
    id                     varchar(255)           not null
        primary key,
    code                   varchar(255)           not null,
    name                   varchar(255)           not null,
    precision              smallint default '8'::smallint not null,
    is_fiat                boolean  default false not null,
    blockchain             varchar(255),
    contract_address       varchar(255),
    withdrawal_min_balance numeric(90, 50),
    has_balance            boolean  default true  not null,
    status                 boolean  default true  not null,
    sort_order             smallint default 1     not null,
    min_confirmation       smallint,
    created_at             timestamp,
    updated_at             timestamp
);

create table if not exists users
(
    id                  uuid         default gen_random_uuid() not null
        primary key,
    email               varchar(255)                           not null
        constraint users_email_unique
            unique,
    email_verified_at   timestamp,
    password            varchar(255)                           not null,
    remember_token      varchar(100),
    processing_owner_id uuid
        constraint users_processing_owner_id_unique
            unique,
    location            varchar(50),
    language            varchar(2)   default 'en'::character varying      not null,
    rate_source         varchar(255) default 'Binance'::character varying not null,
    created_at          timestamp,
    updated_at          timestamp,
    deleted_at          timestamp,
    banned              boolean      default false
);

create table if not exists personal_access_tokens
(
    id             uuid default gen_random_uuid() not null
        primary key,
    tokenable_type varchar(255)                   not null,
    tokenable_id   uuid                           not null,
    name           varchar(255)                   not null,
    token          varchar(64)                    not null,
    last_used_at   timestamp(0),
    expires_at     timestamp(0),
    created_at     timestamp(0),
    updated_at     timestamp(0)
);

create index if not exists personal_access_tokens_tokenable_type_tokenable_id_index
    on personal_access_tokens (tokenable_type, tokenable_id);

create unique index if not exists personal_access_tokens_token_unique
    on personal_access_tokens (token);

create table if not exists stores
(
    id              uuid           default gen_random_uuid() not null
        primary key,
    user_id         uuid                                     not null,
    name            varchar(255)                             not null,
    site            varchar(255),
    currency_id     varchar(255)                             not null,
    rate_source     varchar(255)                             not null,
    return_url      varchar(255),
    success_url     varchar(255),
    rate_scale      numeric(4, 2)  default 1.00              not null,
    status          boolean        default true              not null,
    minimal_payment numeric(28, 2) default 0.10              not null,
    created_at      timestamp,
    updated_at      timestamp,
    deleted_at      timestamp
);

create table if not exists store_api_keys
(
    id         uuid    default gen_random_uuid() not null
        primary key,
    store_id   uuid                              not null
        constraint store_api_keys_store_id_foreign
            references stores,
    key        varchar(255)                      not null
        constraint store_api_keys_key_unique
            unique,
    enabled    boolean default true              not null,
    created_at timestamp,
    updated_at timestamp
);

create table if not exists store_webhooks
(
    id         uuid    default gen_random_uuid() not null
        primary key,
    store_id   uuid                              not null
        constraint webhooks_store_id_foreign
            references stores,
    url        varchar(255)                      not null,
    secret     varchar(255)                      not null,
    enabled    boolean default true              not null,
    events     json                              not null,
    created_at timestamp,
    updated_at timestamp
);

create table if not exists webhook_send_histories
(
    id                   uuid default gen_random_uuid() not null
        primary key,
    tx_id                uuid                           not null,
    send_queue_job_id    uuid                           not null,
    type                 varchar(30)                    not null,
    url                  varchar(255)                   not null,
    status               varchar(255)                   not null,
    request              json,
    response             text,
    response_status_code integer,
    created_at           timestamp                      not null,
    updated_at           timestamp
);

create index if not exists idx_webhook_send_histories_tx_id
    on webhook_send_histories using hash (tx_id);

create index if not exists idx_webhook_send_histories_queue_job_id
    on webhook_send_histories (send_queue_job_id);

create index if not exists idx_webhook_send_histories_created_at
    on webhook_send_histories (created_at);

create table if not exists transactions
(
    id                   uuid    default gen_random_uuid() not null
        primary key,
    user_id              uuid                              not null
        constraint transactions_user_id_foreign
            references users,
    store_id             uuid
        constraint transactions_store_id_foreign
            references stores,
    receipt_id           uuid,
    wallet_id            uuid,
    currency_id          varchar(255)                      not null,
    blockchain           varchar(255)                      not null,
    tx_hash              varchar(255)                      not null,
    bc_uniq_key          varchar(30),
    type                 varchar(30)                       not null,
    from_address         varchar(255),
    to_address           varchar(255)                      not null,
    amount               numeric(90, 50)                   not null,
    amount_usd           numeric(28, 4),
    fee                  numeric(28, 8)                    not null,
    withdrawal_is_manual boolean default false             not null,
    network_created_at   timestamp,
    created_at           timestamp,
    updated_at           timestamp,
    created_at_index     bigint,
    constraint transactions_blockchain_tx_hash_bc_key_unique
        unique (blockchain, tx_hash, bc_uniq_key)
);

create index if not exists transactions_created_at_index_index
    on transactions (created_at_index);

create index if not exists transactions_receipt_id_index
    on transactions (receipt_id);

create index if not exists transactions_store_id_index
    on transactions (store_id);

create index if not exists transactions_to_address_index
    on transactions (to_address);

create index if not exists transactions_tx_id_index
    on transactions (tx_hash);

create index if not exists transactions_from_address_index
    on transactions (from_address);

create table if not exists wallets
(
    id                uuid default gen_random_uuid() not null
        primary key,
    store_id          uuid                           not null
        constraint wallets_store_id_foreign
            references stores,
    store_external_id varchar(255)                   not null,
    created_at        timestamp,
    updated_at        timestamp,
    deleted_at        timestamp
);

create table if not exists wallet_addresses
(
    id          uuid            default gen_random_uuid() not null
        primary key,
    wallet_id   uuid                                      not null
        constraint wallet_addresses_wallet_id_foreign
            references wallets,
    user_id     uuid                                      not null
        constraint wallet_addresses_user_id_foreign
            references users,
    currency_id varchar(255)                              not null
        constraint wallet_addresses_currency_id_foreign
            references currencies,
    blockchain  varchar(255)                              not null,
    address     varchar(255)                              not null,
    amount      numeric(90, 50) default 0                 not null,
    created_at  timestamp,
    updated_at  timestamp,
    deleted_at  timestamp,
    dirty       boolean         default false             not null
);

create table if not exists receipts
(
    id          uuid           default gen_random_uuid() not null
        primary key,
    status      varchar(255)                             not null,
    store_id    uuid                                     not null
        constraint invoices_store_id_foreign
            references stores,
    currency_id varchar(255)                             not null
        constraint invoices_currency_id_foreign
            references currencies,
    amount      numeric(28, 8) default 0.00000000        not null,
    wallet_id   uuid
        constraint invoices_payer_id_foreign
            references wallets,
    created_at  timestamp,
    updated_at  timestamp
);

create table if not exists transfers
(
    id           uuid         default gen_random_uuid() not null
        primary key,
    number       bigserial                              not null
        constraint transfers_number_unique
            unique,
    user_id      uuid                                   not null
        constraint transfers_user_id_foreign
            references users,
    kind         varchar(255) default 'transferFromAddress'::character varying not null,
    currency_id  varchar(255)                           not null
        constraint transfers_currency_id_foreign
            references currencies,
    status       varchar(255)                           not null,
    stage        varchar(255)                           not null,
    from_address varchar(255)                           not null,
    to_address   varchar(255)                           not null,
    amount       numeric(90, 50)                        not null,
    amount_usd   numeric(28, 4)                         not null,
    message      varchar(255),
    created_at   timestamp,
    updated_at   timestamp
);

create index if not exists transfers_kind_index
    on transfers (kind);

create index if not exists transfers_uuid_status_index
    on transfers (id, status);

create index if not exists transfers_stage_index
    on transfers (stage);

create table if not exists withdrawal_wallets
(
    id                     uuid         default gen_random_uuid() not null
        primary key,
    user_id                uuid                                   not null
        constraint withdrawal_wallets_user_id_foreign
            references users,
    blockchain             varchar(255)                           not null,
    currency_id            varchar(255)                           not null,
    withdrawal_min_balance numeric(90, 50)                        not null,
    withdrawal_interval    varchar(255) default 'never'::character varying   not null,
    created_at             timestamp,
    deleted_at             timestamp,
    updated_at             timestamp,
    withdrawal_enabled     varchar(255) default 'enabled'::character varying not null
);

create table if not exists withdrawal_wallet_addresses
(
    id                   uuid default gen_random_uuid() not null
        primary key,
    withdrawal_wallet_id uuid                           not null
        constraint withdrawal_wallet_addresses_withdrawal_wallet_id_foreign
            references withdrawal_wallets,
    name                 varchar(255),
    address              varchar(255)                   not null,
    created_at           timestamp                      not null,
    updated_at           timestamp,
    deleted_at           timestamp,
    unique (withdrawal_wallet_id, address)
);

create table if not exists settings
(
    id         uuid default gen_random_uuid() not null
        primary key,
    model_id   uuid,
    model_type varchar(255),
    name       varchar(255)                   not null,
    value      varchar(255)                   not null,
    created_at timestamp,
    updated_at timestamp
);

create unique index if not exists unique_name_when_model_is_null
    on settings (name)
    where ((model_id IS NULL) AND (model_type IS NULL));

create table if not exists store_currencies
(
    currency_id varchar(255) not null,
    store_id    uuid         not null
);

create index if not exists currency_store_currency_id_store_id_index
    on store_currencies (currency_id, store_id);

create table if not exists store_whitelist
(
    ip       varchar(255) not null,
    store_id uuid         not null
);

create index if not exists currency_store_whitelist_store_id_index
    on store_whitelist (ip, store_id);

create table if not exists unconfirmed_transactions
(
    id                 uuid default gen_random_uuid() not null
        primary key,
    user_id            uuid                           not null
        constraint unconfirmed_transactions_user_id_foreign
            references users,
    store_id           uuid
        constraint unconfirmed_transactions_store_id_foreign
            references stores,
    wallet_id          uuid,
    currency_id        varchar(255)                   not null,
    tx_hash            varchar(255)                   not null,
    bc_uniq_key        varchar(30),
    type               varchar(30)                    not null,
    from_address       varchar(255),
    to_address         varchar(255)                   not null,
    amount             numeric(90, 50)                not null,
    amount_usd         numeric(28, 4),
    network_created_at timestamp,
    created_at         timestamp,
    updated_at         timestamp,
    constraint unconfirmed_transactions_tx_id_bc_key_unique
        unique (tx_hash, bc_uniq_key)
);

create index if not exists idx_unconfirmed_transactions_to_address
    on unconfirmed_transactions (to_address);

create index if not exists unconfirmed_transactions_store_id_index
    on unconfirmed_transactions (store_id);

create index if not exists unconfirmed_transactions_to_address_index
    on unconfirmed_transactions (to_address);

create index if not exists unconfirmed_transactions_tx_id_index
    on unconfirmed_transactions (tx_hash);

create table if not exists webhook_send_queue
(
    id             uuid         default gen_random_uuid() not null
        primary key,
    webhook_id     uuid                                   not null,
    seconds_delay  smallint                               not null,
    transaction_id uuid                                   not null,
    payload        json                                   not null,
    signature      varchar(255)                           not null,
    event          varchar(50)                            not null,
    last_sent_at   timestamp(0) default NULL::timestamp without time zone,
    created_at     timestamp(0),
    constraint uniq_webhook_id_tx_id
        unique (webhook_id, transaction_id)
);

create table if not exists exchanges
(
    id         uuid    default gen_random_uuid() not null
        primary key,
    slug       varchar(255)                      not null
        constraint slug_unique
            unique,
    name       varchar(255)                      not null,
    is_active  boolean default false             not null,
    url        varchar(255)                      not null,
    created_at timestamp                         not null,
    updated_at timestamp                         not null
);

create table if not exists exchange_keys
(
    id          uuid default gen_random_uuid() not null
        primary key,
    exchange_id uuid                           not null,
    name        varchar(255)                   not null,
    title       varchar(255)                   not null,
    created_at  timestamp,
    updated_at  timestamp
);

create index if not exists exchange_id_index
    on exchange_keys (exchange_id);

create table if not exists exchange_user_keys
(
    id              uuid default gen_random_uuid() not null
        primary key,
    user_id         uuid                           not null,
    exchange_key_id uuid                           not null,
    value           text                           not null,
    created_at      timestamp,
    updated_at      timestamp,
    constraint user_exchange_key_uniq
        unique (user_id, exchange_key_id)
);

create index if not exists user_id_exchange_key_id_idx
    on exchange_user_keys (user_id, exchange_key_id);

