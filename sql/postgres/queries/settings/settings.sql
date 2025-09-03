-- name: GetByName :one
SELECT *
FROM settings
WHERE name = $1
  and model_id is null
  and model_type is null
LIMIT 1;

-- name: GetByNames :many
SELECT *
FROM settings
WHERE name = ANY (sqlc.arg(setting_names)::varchar[])
  and model_id is null;

-- name: GetByModel :one
SELECT *
FROM settings
WHERE name = $1
  and model_id = $2
  and model_type = $3
LIMIT 1;

-- name: GetAllRootSettings :many
SELECT *
FROM settings
WHERE model_id is null
  and model_type is null;

-- name: GetAllByModel :many
SELECT *
FROM settings
WHERE model_id = $1
  and model_type = $2;

-- name: UpdateSettingsWithReplace :batchexec
INSERT INTO settings (name, value, model_id, model_type, created_at, updated_at)
VALUES ($1, $2, $3, $4, now(), now())
ON CONFLICT (name, model_id, model_type) DO UPDATE
    SET value      = $2,
        updated_at = now();

-- name: DeleteRootByName :exec
DELETE FROM settings WHERE name=$1;

-- name: DeleteByNameAndModelID :exec
DELETE FROM settings WHERE name=$1 AND model_id=$2;