-- name: CreateOrUpdateUserKey :one
INSERT INTO exchange_user_keys (user_id, exchange_key_id, value, created_at, updated_at)
 VALUES (sqlc.arg(user_id), sqlc.arg(exchange_key_id), sqlc.arg(value), now(), now())
ON CONFLICT (user_id, exchange_key_id) DO UPDATE
 SET value = sqlc.arg(value),
     updated_at = now()
RETURNING *;

-- name: DeleteByID :exec
DELETE FROM exchange_user_keys WHERE id=sqlc.arg(id);

-- name: BatchDeleteByID :batchexec
DELETE FROM exchange_user_keys WHERE id = $1;

-- name: GetKeysByExchangeSlug :many
SELECT euk.id, ek.name, euk.value FROM exchanges e
    JOIN exchange_keys ek ON e.id = ek.exchange_id
    JOIN exchange_user_keys euk ON ek.id = euk.exchange_key_id AND euk.user_id = sqlc.arg(user_id)
where e.slug = sqlc.arg(exchange_slug);