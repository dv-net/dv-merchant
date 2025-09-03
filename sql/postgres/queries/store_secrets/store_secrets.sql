-- name: GetSecretByStoreID :one
SELECT secret FROM store_secrets WHERE store_id = $1 LIMIT 1;

-- name: DeleteBuStoreID :exec
DELETE FROM store_secrets WHERE store_id=$1;

-- name: Create :one
INSERT INTO store_secrets (store_id, secret)
VALUES ($1, $2) ON CONFLICT (store_id) DO UPDATE SET secret = $2, updated_at = now()
RETURNING *;

-- name: UpdateSecret :one
UPDATE store_secrets
SET secret=$1, updated_at=now()
WHERE store_id=$2
    RETURNING *;
