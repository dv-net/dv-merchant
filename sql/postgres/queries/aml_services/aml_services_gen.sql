-- name: GetByID :one
SELECT * FROM aml_services WHERE id=$1 LIMIT 1;

