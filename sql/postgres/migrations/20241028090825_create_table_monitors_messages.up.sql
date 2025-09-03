CREATE TABLE IF NOT EXISTS monitors_messages (
    id uuid PRIMARY KEY default gen_random_uuid(),
    monitor_id uuid NOT NULL,
    failure boolean NOT NULL DEFAULT FALSE,
    level varchar(100) NOT NULL,
    message text NOT NULL,
    created_at TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP
)