-- name: Create :one
INSERT INTO aml_user_keys (key_id, user_id, value, created_at, updated_at)
	VALUES ($1, $2, $3, now(), $4)
	RETURNING *;

