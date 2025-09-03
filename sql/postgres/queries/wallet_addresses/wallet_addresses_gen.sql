-- name: Create :one
INSERT INTO wallet_addresses (wallet_id, user_id, currency_id, blockchain, address, created_at, dirty)
	VALUES ($1, $2, $3, $4, $5, now(), $6)
	RETURNING *;

-- name: GetById :one
SELECT * FROM wallet_addresses WHERE deleted_at IS NULL AND id=$1 LIMIT 1;

