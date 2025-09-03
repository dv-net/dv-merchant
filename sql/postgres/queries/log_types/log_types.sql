-- name: GetBySlug :one
SELECT * FROM log_types WHERE slug = $1 LIMIT 1;

-- name: GetAll :many
SELECT * FROM log_types;
