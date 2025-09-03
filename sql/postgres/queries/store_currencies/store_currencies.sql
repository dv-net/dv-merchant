-- name: Delete :exec
DELETE FROM store_currencies WHERE store_id=$1 and currency_id=$2;

-- name: GetAllByStoreID :many
SELECT * FROM currencies
WHERE id IN (
    SELECT currency_id FROM store_currencies
    WHERE store_id = $1
);

-- name: CreateOne :exec
INSERT INTO store_currencies (currency_id, store_id)
VALUES ($1, $2) ON CONFLICT DO NOTHING;

-- name: FindAllByStoreID :many
SELECT * FROM store_currencies
WHERE store_id=$1;

-- name: FindByStoreID :one
SELECT * FROM store_currencies
WHERE store_id=$1 and currency_id=$2;