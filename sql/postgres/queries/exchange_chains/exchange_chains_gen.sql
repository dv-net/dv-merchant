-- name: GetByID :one
SELECT * FROM exchange_chains WHERE id=$1 LIMIT 1;

