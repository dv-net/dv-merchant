-- name: Create :one
INSERT INTO aml_service_keys (service_id, name, description, created_at, updated_at)
	VALUES ($1, $2, $3, now(), $4)
	RETURNING *;

