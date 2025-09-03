CREATE TABLE IF NOT EXISTS notification_send_history(
    id uuid default gen_random_uuid() primary key,
    destination varchar not null,
    message_text text,
    sender varchar not null,
    created_at timestamp not null default now(),
    updated_at timestamp,
    sent_at timestamp,
    type varchar not null,
    channel varchar not null,
    attempts integer not null default 0
);

CREATE INDEX IF NOT EXISTS notification_send_history_index ON notification_send_history(destination, type, channel);