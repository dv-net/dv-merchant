-- name: Create :exec
INSERT INTO aml_check_queue (user_id, aml_check_id, created_at, request_payload)
	VALUES ($1, $2, now(), $3);
