-- name: GetByStoreID :one
SELECT *
FROM store_aml_settings
WHERE store_id = $1
limit 1;

-- name: Upsert :one
INSERT INTO store_aml_settings (store_id, enabled, risk_threshold, provider_slug, created_at, updated_at)
VALUES ($1, $2, $3, $4, now(), now())
ON CONFLICT (store_id) DO UPDATE
    SET enabled        = EXCLUDED.enabled,
        risk_threshold = EXCLUDED.risk_threshold,
        provider_slug  = EXCLUDED.provider_slug,
        updated_at     = now()
RETURNING *;
