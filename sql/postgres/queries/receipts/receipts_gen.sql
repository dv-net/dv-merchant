-- name: Create :one
INSERT INTO receipts (status, store_id, currency_id, amount, wallet_id, created_at)
	VALUES ($1, $2, $3, $4, $5, now())
	RETURNING *;

-- name: GetAll :many
SELECT * FROM receipts ORDER BY created_at DESC LIMIT $1 OFFSET $2;

-- name: GetByID :one
SELECT * FROM receipts WHERE id=$1 LIMIT 1;

