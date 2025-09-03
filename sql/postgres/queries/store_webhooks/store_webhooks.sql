-- name: GetByStoreId :many
SELECT * FROM store_webhooks WHERE store_id=$1 ORDER BY created_at;

-- name: GetByStoreAndType :many
SELECT sqlc.embed(sw), ss.secret
FROM store_webhooks sw
LEFT JOIN store_secrets ss on ss.store_id = $1
WHERE sw.store_id = $1
  and sqlc.arg(event_type)::varchar in (select json_array_elements_text(sw.events))
  and sw.enabled = true
ORDER BY sw.created_at;

-- name: DisableAllByStore :exec
UPDATE store_webhooks SET enabled = false, updated_at = now() WHERE store_id = $1;

-- name: EnableAllByStore :exec
UPDATE store_webhooks SET enabled = true, updated_at = now() WHERE store_id = $1;
