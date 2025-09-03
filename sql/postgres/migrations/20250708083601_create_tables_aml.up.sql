create table if not exists aml_services
(
    id         uuid primary key      DEFAULT gen_random_uuid(),
    slug       varchar(255) not null check (slug != ''),
    created_at TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP             DEFAULT NULL,

    UNIQUE (slug)
);

create table if not exists aml_service_keys
(
    id          uuid primary key      DEFAULT gen_random_uuid(),
    service_id  uuid         not null references aml_services,
    name        varchar(255) not null check (name != ''),
    description varchar(255) not null check (description != ''),
    created_at  TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP             DEFAULT NULL,

    UNIQUE (service_id, name)
);

create table if not exists aml_user_keys
(
    id         uuid primary key      DEFAULT gen_random_uuid(),
    key_id     uuid         not null references aml_service_keys,
    user_id    uuid         not null references users,
    value      varchar(255) not null check (value != ''),
    created_at TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP             DEFAULT NULL,

    UNIQUE (user_id, key_id)
);

create table if not exists aml_checks
(
    id          uuid primary key       DEFAULT gen_random_uuid(),
    user_id     uuid          not null references users,
    service_id  uuid          not null references aml_services,
    external_id varchar(255)  not null check ( external_id != '' ),
    status      varchar(255)  not null check (status != ''),
    score       NUMERIC(5, 3) not null default 0 check (score >= 0),
    risk_level  varchar(255)           default null,
    created_at  TIMESTAMP     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP              DEFAULT NULL,

    UNIQUE (external_id, service_id)
);

CREATE INDEX idx_aml_checks_status_service_id ON aml_checks (service_id, status);

CREATE TABLE IF NOT EXISTS aml_check_history
(
    id               UUID PRIMARY KEY   DEFAULT gen_random_uuid(),
    aml_check_id     UUID      NOT NULL REFERENCES aml_checks (id),
    request_payload  JSONB     NOT NULL DEFAULT '{}',
    service_response JSONB     NOT NULL DEFAULT '{}',
    error_msg VARCHAR(255) DEFAULT NULL,
    attempt_number   INTEGER   NOT NULL,
    created_at       TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at       TIMESTAMP          DEFAULT NULL
);

create table if not exists aml_check_queue
(
    id           uuid primary key DEFAULT gen_random_uuid(),
    user_id      uuid      not null references users,
    aml_check_id uuid      not null references aml_checks,
    attempts     integer   not null default 0,
    created_at   TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at   TIMESTAMP          DEFAULT NULL,
    UNIQUE (aml_check_id)
);

create table if not exists aml_supported_assets (
    service_slug varchar(255) not null check(service_slug != ''),
    currency_id varchar(255) not null check(currency_id != ''),
    asset_identity varchar(255) not null check (asset_identity != ''),
    blockchain_name varchar(255) not null check(blockchain_name != ''),
    UNIQUE(service_slug, currency_id)
);