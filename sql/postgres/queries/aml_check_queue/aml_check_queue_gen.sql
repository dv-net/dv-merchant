-- name: Create :exec
INSERT INTO aml_check_queue (user_id, aml_check_id, created_at)
	VALUES ($1, $2, now());

