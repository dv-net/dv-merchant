-- name: ListRiskRulesByUserID :many
SELECT *
FROM user_aml_risk_rules
WHERE user_id = $1
  AND provider_slug = $2
ORDER BY risk_type;

-- name: UpsertAmlRiskRule :one
INSERT INTO user_aml_risk_rules (user_id, provider_slug, risk_type, enabled, threshold, action,
                                 created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, now(), now()) ON CONFLICT (user_id, provider_slug, risk_type) DO
UPDATE
    SET enabled = EXCLUDED.enabled,
    threshold = EXCLUDED.threshold,
    action = EXCLUDED.action,
    updated_at = now()
    RETURNING *;