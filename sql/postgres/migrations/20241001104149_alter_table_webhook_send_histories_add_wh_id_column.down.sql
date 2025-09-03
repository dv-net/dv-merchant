ALTER TABLE webhook_send_histories
    DROP COLUMN IF EXISTS is_manual;
ALTER TABLE webhook_send_histories
    ALTER COLUMN send_queue_job_id SET not null;