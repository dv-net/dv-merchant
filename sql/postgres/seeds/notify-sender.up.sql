INSERT INTO settings (name, value, created_at, is_mutable)
VALUES ('notification_sender', 'dv_net', NOW(), true)
ON CONFLICT DO NOTHING;