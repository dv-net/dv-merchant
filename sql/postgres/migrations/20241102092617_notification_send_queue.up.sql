CREATE TABLE IF NOT EXISTS notification_send_queue
(
    id          uuid      default gen_random_uuid() not null
        primary key,
    destination varchar(255)                        not null CHECK ( notification_send_queue.destination != ''),
    type        varchar                             not null,
    parameters  jsonb                               not null,
    channel     varchar(255)                        not null,
    attempts    integer   default 0                 not null,
    created_at  timestamp default now()             not null,
    updated_at  timestamp
);

ALTER TABLE IF EXISTS notification_send_history
DROP COLUMN IF EXISTS attempts;

TRUNCATE TABLE notification_send_history;

ALTER TABLE IF EXISTS notification_send_history
    ADD COLUMN IF NOT EXISTS notification_send_queue_id uuid not null;