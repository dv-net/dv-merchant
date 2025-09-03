INSERT INTO settings (name, value, created_at, is_mutable) VALUES
    ('mailer_state', 'disabled', NOW(), true),
    ('mailer_protocol', 'smtp', NOW(), false),
    ('mailer_address', 'localhost:1025', NOW(), true),
    ('mailer_username', '', NOW(), true),
    ('mailer_password', '', NOW(), true),
    ('mailer_sender', 'admin@example.com', NOW(), true),
    ('mailer_encryption', 'NONE', NOW(), true)
ON CONFLICT DO NOTHING;