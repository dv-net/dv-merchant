DROP INDEX IF EXISTS webhook_send_queue_created_at_idx;

ALTER TABLE webhook_send_queue
    ALTER COLUMN seconds_delay TYPE smallint USING least(seconds_delay, 32767)::smallint;
