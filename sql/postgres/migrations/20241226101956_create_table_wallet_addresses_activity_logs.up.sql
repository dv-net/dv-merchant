CREATE TABLE IF NOT EXISTS wallet_addresses_activity_logs
(
    id                  uuid PRIMARY KEY      DEFAULT gen_random_uuid(),
    wallet_addresses_id uuid         NOT NULL,
    text                VARCHAR(255) NOT NULL,
    text_variables      JSONB        NOT NULL DEFAULT '{}',
    created_at          timestamp             default now(),
    updated_at          timestamp
);
