-- name: UpdateAMLCheck :exec
UPDATE aml_checks
SET status     = $2,
    score      = $3,
    risk_level = $4,
    updated_at = now()
WHERE id = $1;