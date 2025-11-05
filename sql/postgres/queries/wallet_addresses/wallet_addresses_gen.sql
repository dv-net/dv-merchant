-- name: Create :one
INSERT INTO wallet_addresses (user_id, currency_id, blockchain, address, created_at, dirty, status, account_type, account_id, store_id)
	VALUES ($1, $2, $3, $4, now(), $5, $6, $7, $8, $9)
	RETURNING *;

-- name: GetById :one
SELECT * FROM wallet_addresses WHERE deleted_at IS NULL AND id=$1 LIMIT 1;

