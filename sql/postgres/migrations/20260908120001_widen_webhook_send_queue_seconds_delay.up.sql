ALTER TABLE webhook_send_queue
    ALTER COLUMN seconds_delay TYPE integer;

CREATE INDEX IF NOT EXISTS webhook_send_queue_created_at_idx
    ON webhook_send_queue (created_at);
