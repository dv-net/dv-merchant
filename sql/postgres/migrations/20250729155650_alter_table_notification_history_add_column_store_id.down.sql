ALTER TABLE notification_send_history DROP COLUMN IF EXISTS store_id;
ALTER TABLE notification_send_history DROP COLUMN IF EXISTS user_id;
ALTER TABLE notification_send_queue DROP COLUMN IF EXISTS args;