-- name: Create :one
INSERT INTO exchange_addresses (address, chain, currency, address_type, user_id, create_type, created_at, exchange_id)
	VALUES ($1, $2, $3, $4, $5, $6, now(), $7)
	RETURNING *;

-- name: GetAllByUser :many
SELECT * FROM exchange_addresses WHERE exchange_id=$1 AND user_id=$2;

