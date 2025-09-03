-- name: Create :one
INSERT INTO store_webhooks (store_id, url, enabled, events, created_at)
	VALUES ($1, $2, $3, $4, now())
	RETURNING *;

-- name: Delete :exec
DELETE FROM store_webhooks WHERE id=$1;

-- name: GetById :one
SELECT * FROM store_webhooks WHERE id=$1 LIMIT 1;

-- name: Update :one
UPDATE store_webhooks
	SET url=$1, enabled=$2, events=$3, updated_at=$4
	WHERE id=$5
	RETURNING *;

