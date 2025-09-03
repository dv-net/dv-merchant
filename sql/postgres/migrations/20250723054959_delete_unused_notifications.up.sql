DELETE FROM notifications
WHERE type NOT IN (
	'user_verification',
	'user_registration',
	'user_password_reset',
	'user_forgot_password',
	'two_factor_authentication',
	'external_wallet_requested',
	'user_invite',
	'user_email_reset',
	'user_remind_verification',
	'user_update_setting_verification',
	'user_change_email',
	'user_authorization_from_new_device',
	'user_access_key_changed'
) AND category IS NULL;