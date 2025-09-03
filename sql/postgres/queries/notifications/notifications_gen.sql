-- name: Create :one
INSERT INTO notifications (category, type)
	VALUES ($1, $2)
	RETURNING *;

-- name: GetByID :one
SELECT * FROM notifications WHERE id=$1 LIMIT 1;

