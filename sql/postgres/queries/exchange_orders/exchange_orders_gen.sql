-- name: Create :one
INSERT INTO exchange_orders (exchange_id, exchange_order_id, client_order_id, symbol, side, amount, order_created_at, created_at, fail_reason, status, user_id, amount_usd, exchange_connection_hash)
	VALUES ($1, $2, $3, $4, $5, $6, $7, now(), $8, $9, $10, $11, $12)
	RETURNING *;

-- name: GetByID :one
SELECT * FROM exchange_orders WHERE id=$1 LIMIT 1;

