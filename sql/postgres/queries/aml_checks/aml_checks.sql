-- name: UpdateAMLCheck :exec
UPDATE aml_checks
SET status     = $2,
    score      = $3,
    risk_level = $4,
    updated_at = now()
WHERE id = $1;

-- name: GetByTransactionID :one
SELECT *
FROM aml_checks
WHERE transaction_id = $1
LIMIT 1;

-- name: UpdateExternalID :exec
UPDATE aml_checks
SET external_id = $2,
    updated_at  = now()
WHERE id = $1;

-- name: GetStatistics :one
SELECT
    COUNT(*) FILTER (WHERE status IN ('success', 'failed')) AS checked_today,
    COUNT(*) FILTER (WHERE status = 'success') AS successful_today,
    COUNT(*) FILTER (WHERE status = 'failed') AS failed_today
FROM aml_checks
WHERE user_id = sqlc.arg(user_id)
  AND created_at >= sqlc.arg(date_from)::timestamp
  AND created_at < sqlc.arg(date_to)::timestamp;
