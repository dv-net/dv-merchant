-- name: GetByStoreId :many
SELECT * FROM store_webhooks WHERE store_id=$1 ORDER BY created_at;

-- name: GetByStoreAndType :many
SELECT sqlc.embed(sw), ss.secret
FROM store_webhooks sw
LEFT JOIN store_secrets ss on ss.store_id = $1
WHERE sw.store_id = $1
  AND sw.events::jsonb @> jsonb_build_array(sqlc.arg(event_type)::text)
  and sw.enabled = true
ORDER BY sw.created_at;

-- name: DisableAllByStore :exec
UPDATE store_webhooks SET enabled = false, updated_at = now() WHERE store_id = $1;

-- name: EnableAllByStore :exec
UPDATE store_webhooks SET enabled = true, updated_at = now() WHERE store_id = $1;
