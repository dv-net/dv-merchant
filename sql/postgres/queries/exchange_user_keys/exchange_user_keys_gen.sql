-- name: Create :one
INSERT INTO exchange_user_keys (user_id, exchange_key_id, value, created_at)
	VALUES ($1, $2, $3, now())
	RETURNING *;

