-- name: GetByStoreId :many
SELECT *
FROM store_api_keys
WHERE store_id =$1
ORDER BY created_at;

-- name: GetStoreByKey :one
SELECT s.*
FROM store_api_keys sak
         INNER JOIN stores s on s.id = sak.store_id
WHERE sak.key = $1 and sak.enabled = true
LIMIT 1;

-- name: UpdateKey :one
UPDATE store_api_keys
SET key=$1, updated_at=now()
WHERE store_id=$2
    RETURNING *;

-- name: DisableByStore :exec
UPDATE store_api_keys SET enabled=false, updated_at = now() WHERE store_id = $1;

-- name: EnableByStore :exec
UPDATE store_api_keys SET enabled=true, updated_at = now() WHERE store_id = $1;