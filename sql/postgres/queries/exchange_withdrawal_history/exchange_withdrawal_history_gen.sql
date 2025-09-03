-- name: Create :one
INSERT INTO exchange_withdrawal_history (user_id, exchange_id, exchange_order_id, address, native_amount, fiat_amount, currency, chain, status, created_at, fail_reason, exchange_connection_hash)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, now(), $10, $11)
	RETURNING *;

