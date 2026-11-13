INSERT INTO notifications (category, type)
VALUES (null, 'deposit_aml_blocked')
ON CONFLICT DO NOTHING;

INSERT INTO
	settings (model_id, model_type, name, value, created_at, is_mutable)
SELECT
	id,
	'Store',
	'deposit_aml_blocked_email_notification',
	'enabled',
	NOW (),
	true
FROM
	stores ON CONFLICT DO NOTHING;
