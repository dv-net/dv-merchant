-- name: GetByUserID :one
SELECT *
FROM user_aml_settings
WHERE user_id = $1 limit 1;

-- name: UpsertAmlSetting :one
INSERT INTO user_aml_settings (user_id, enabled, provider_slug, created_at, updated_at)
VALUES ($1, $2, $3, now(), now()) ON CONFLICT (user_id) DO
UPDATE
    SET enabled = EXCLUDED.enabled,
    provider_slug = EXCLUDED.provider_slug,
    updated_at = now()
    RETURNING *;

