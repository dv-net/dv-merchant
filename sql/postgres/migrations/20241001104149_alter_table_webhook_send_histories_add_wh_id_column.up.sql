ALTER TABLE webhook_send_histories
    ALTER COLUMN send_queue_job_id DROP not null;
ALTER TABLE webhook_send_histories
    ADD COLUMN IF NOT EXISTS is_manual boolean DEFAULT false;