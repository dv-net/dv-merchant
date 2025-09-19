INSERT INTO
	settings (model_id, model_type, name, value, created_at, is_mutable)
SELECT
	id,
	'Store',
	'external_wallet_email_notification',
	'enabled',
	NOW (),
	true
FROM
	stores
UNION ALL
SELECT
	id,
	'Store',
	'user_crypto_receipt_email_notification',
	'enabled',
	NOW (),
	true
FROM
	stores ON CONFLICT DO NOTHING;