-- name: Create :one
INSERT INTO exchanges (slug, name, is_active, url, created_at)
	VALUES ($1, $2, $3, $4, now())
	RETURNING *;

-- name: GetAll :many
SELECT * FROM exchanges;

-- name: GetByID :one
SELECT * FROM exchanges WHERE id=$1 LIMIT 1;

