-- name: Create :one
INSERT INTO user_stores (user_id, store_id, created_at)
	VALUES ($1, $2, now())
	RETURNING *;

