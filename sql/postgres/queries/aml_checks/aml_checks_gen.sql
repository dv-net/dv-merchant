-- name: Create :one
INSERT INTO aml_checks (user_id, service_id, external_id, status, score, risk_level, created_at, updated_at, transaction_id)
	VALUES ($1, $2, $3, $4, $5, $6, now(), $7, $8)
	RETURNING *;
