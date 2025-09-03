-- name: Create :one
INSERT INTO aml_check_history (aml_check_id, request_payload, service_response, error_msg, attempt_number, created_at)
	VALUES ($1, $2, $3, $4, $5, now())
	RETURNING *;

