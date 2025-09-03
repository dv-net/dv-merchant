-- name: Create :one
INSERT INTO settings (model_id, model_type, name, value, created_at, is_mutable)
	VALUES ($1, $2, $3, $4, now(), $5)
	RETURNING *;

-- name: Update :one
UPDATE settings
	SET model_id=$1, model_type=$2, name=$3, value=$4, updated_at=$5, is_mutable=$6
	WHERE id=$7
	RETURNING *;

