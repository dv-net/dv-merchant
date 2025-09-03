CREATE TABLE IF NOT EXISTS store_secrets
(
    id         uuid    default gen_random_uuid() not null
        primary key,
    store_id   uuid                              not null
        constraint store_secrets_store_id_foreign
            references stores unique,
    secret        varchar(255)                      not null,
    created_at timestamp not null default now(),
    updated_at timestamp
);

INSERT INTO store_secrets (store_id, secret, created_at) SELECT store_id, (
    SELECT secret
    FROM store_webhooks sw2
    WHERE sw2.store_id = sw1.store_id
    ORDER BY random()
    LIMIT 1
) AS secret, now()
FROM store_webhooks sw1
GROUP BY store_id;

ALTER TABLE IF EXISTS store_webhooks DROP COLUMN IF EXISTS secret;