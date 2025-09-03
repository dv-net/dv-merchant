-- name: Create :one
INSERT INTO withdrawal_wallet_addresses (withdrawal_wallet_id, name, address, created_at)
	VALUES ($1, $2, $3, now())
	RETURNING *;

-- name: GetById :one
SELECT * FROM withdrawal_wallet_addresses WHERE deleted_at IS NULL AND id=$1 LIMIT 1;

