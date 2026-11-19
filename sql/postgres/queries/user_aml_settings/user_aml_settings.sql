-- name: GetByUserID :one
SELECT *
FROM user_aml_settings
WHERE user_id = $1 limit 1;

-- name: UpsertAmlSetting :one
INSERT INTO user_aml_settings (user_id, enabled, provider_slug, min_check_amount_usd, created_at, updated_at)
VALUES (sqlc.arg(user_id), sqlc.arg(enabled), sqlc.arg(provider_slug),
        COALESCE(sqlc.narg(min_check_amount_usd)::numeric, 0), now(), now()) ON CONFLICT (user_id) DO
UPDATE
    SET enabled = EXCLUDED.enabled,
    provider_slug = EXCLUDED.provider_slug,
    min_check_amount_usd = COALESCE(sqlc.narg(min_check_amount_usd)::numeric, user_aml_settings.min_check_amount_usd),
    updated_at = now()
    RETURNING *;
