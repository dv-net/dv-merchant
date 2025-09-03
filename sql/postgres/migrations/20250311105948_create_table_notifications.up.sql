create table if not exists notifications
(
    id        uuid PRIMARY KEY      DEFAULT gen_random_uuid(),
    category  varchar(255)          default null,
    type      varchar(255) not null
);

create table if not exists user_notifications
(
    id              uuid PRIMARY KEY   DEFAULT gen_random_uuid(),
    user_id         uuid      not null,
    notification_id uuid      not null,
    email_enabled   bool      not null default false,
    tg_enabled      bool      not null default false,
    UNIQUE (user_id, notification_id),
    created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP NULL     DEFAULT CURRENT_TIMESTAMP
);

