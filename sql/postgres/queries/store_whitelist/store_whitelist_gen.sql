-- name: Create :one
INSERT INTO store_whitelist (ip, store_id)
	VALUES ($1, $2)
	RETURNING *;

-- name: Find :many
SELECT * FROM store_whitelist WHERE store_id=$1;

