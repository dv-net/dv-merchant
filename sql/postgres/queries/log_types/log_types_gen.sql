-- name: Create :one
INSERT INTO log_types (slug, title, error_count, error_count_notify_limit, start_params, notify_params, created_at)
	VALUES ($1, $2, $3, $4, $5, $6, now())
	RETURNING *;

-- name: GetByID :one
SELECT * FROM log_types WHERE id=$1 LIMIT 1;

-- name: Update :one
UPDATE log_types
	SET slug=$1, title=$2, error_count=$3, error_count_notify_limit=$4, start_params=$5, notify_params=$6, 
		updated_at=now()
	WHERE id=$7
	RETURNING *;

