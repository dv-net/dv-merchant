DELETE FROM settings WHERE name = 'deposit_aml_blocked_email_notification';

DELETE FROM notifications WHERE type = 'deposit_aml_blocked';
