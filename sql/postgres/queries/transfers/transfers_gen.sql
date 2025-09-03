-- name: Create :one
INSERT INTO transfers (id, user_id, kind, currency_id, status, stage, amount, amount_usd, message, created_at, updated_at, blockchain, from_addresses, to_addresses, step, tx_hash)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, now(), now(), $10, $11, $12, $13, $14)
	RETURNING *;

-- name: GetAll :many
SELECT * FROM transfers ORDER BY created_at DESC LIMIT $1 OFFSET $2;

-- name: GetById :one
SELECT * FROM transfers WHERE id=$1 LIMIT 1;

