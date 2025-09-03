-- name: Create :one
INSERT INTO store_api_keys (store_id, key, enabled, created_at)
	VALUES ($1, $2, $3, $4)
	RETURNING *;

-- name: Delete :exec
DELETE FROM store_api_keys WHERE id=$1;

-- name: GetById :one
SELECT * FROM store_api_keys WHERE id=$1 LIMIT 1;

-- name: UpdateStatus :one
UPDATE store_api_keys
	SET enabled=$1, updated_at=$2
	WHERE id=$3
	RETURNING *;

