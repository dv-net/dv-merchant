-- name: GetByUser :many
SELECT *
FROM stores
WHERE user_id = $1
  AND deleted_at IS NULL
ORDER BY created_at DESC;

-- name: GetStoreByStoreApiKey :one
SELECT s.*
FROM stores s
         JOIN
     store_api_keys sak
     ON
         s.id = sak.store_id
WHERE s.deleted_at IS NULL
  AND sak.key = $1
  AND sak.enabled = true
LIMIT 1;

-- name: GetStoreByWalletAddress :one
SELECT s.*
FROM stores s
         LEFT JOIN wallet_addresses wa
                   ON wa.store_id = s.id
WHERE wa.address = $1
  AND wa.currency_id = $2
LIMIT 1;

-- name: GetStoreCurrencies :many
SELECT c.*
FROM store_currencies sc
         INNER JOIN currencies c on sc.currency_id = c.id
WHERE sc.store_id = $1;

-- name: UpdateRate :exec
UPDATE stores
SET rate_scale=$1,
    rate_source = $2,
    updated_at=now()
WHERE user_id = $3;

-- name: GetByIDWithPublicFormEnabled :one
SELECT *
FROM stores
WHERE id = sqlc.arg(store_id)::uuid
  AND public_payment_form_enabled = true
LIMIT 1;

-- name: SoftDelete :exec
UPDATE stores
SET deleted_at = now(), updated_at = now()
WHERE id = $1;

-- name: Restore :exec
UPDATE stores
SET deleted_at = NULL, updated_at = now()
WHERE id = $1;

-- name: GetArchivedByUser :many
SELECT *
FROM stores
WHERE user_id = $1
  AND deleted_at IS NOT NULL
ORDER BY created_at DESC;

-- name: GetStoreByWalletID :one
SELECT s.*
FROM stores s 
JOIN wallets w ON s.id = w.store_id
WHERE w.id = $1
LIMIT 1;