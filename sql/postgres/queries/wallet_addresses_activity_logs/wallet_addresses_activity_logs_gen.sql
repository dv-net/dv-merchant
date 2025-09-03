-- name: Create :one
INSERT INTO wallet_addresses_activity_logs (wallet_addresses_id, text, text_variables, created_at)
	VALUES ($1, $2, $3, now())
	RETURNING *;

