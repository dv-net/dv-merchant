ALTER TABLE notification_send_history ADD COLUMN IF NOT EXISTS store_id UUID NULL;
ALTER TABLE notification_send_history ADD COLUMN IF NOT EXISTS user_id UUID NULL;
ALTER TABLE notification_send_queue ADD COLUMN IF NOT EXISTS args JSONB NULL;