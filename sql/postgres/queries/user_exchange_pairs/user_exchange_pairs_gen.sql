-- name: Create :one
INSERT INTO user_exchange_pairs (exchange_id, user_id, currency_from, currency_to, symbol, type)
	VALUES ($1, $2, $3, $4, $5, $6)
	RETURNING *;

-- name: Find :many
SELECT * FROM user_exchange_pairs WHERE exchange_id=$1 AND user_id=$2;

