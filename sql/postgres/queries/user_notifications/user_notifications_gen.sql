-- name: Create :one
INSERT INTO user_notifications (user_id, notification_id, email_enabled, tg_enabled, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5, $6)
	RETURNING *;

-- name: GetByID :one
SELECT * FROM user_notifications WHERE id=$1 LIMIT 1;

