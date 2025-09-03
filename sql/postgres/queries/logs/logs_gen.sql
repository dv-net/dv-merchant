-- name: Create :one
INSERT INTO logs (log_type_slug, process_id, level, status, message, created_at)
	VALUES ($1, $2, $3, $4, $5, now())
	RETURNING *;

-- name: GetByID :one
SELECT * FROM logs WHERE id=$1 LIMIT 1;

-- name: Update :one
UPDATE logs
	SET log_type_slug=$1, process_id=$2, level=$3, status=$4, message=$5
	WHERE id=$6
	RETURNING *;

