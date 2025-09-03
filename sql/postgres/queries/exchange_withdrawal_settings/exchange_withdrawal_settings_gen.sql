-- name: Create :one
INSERT INTO exchange_withdrawal_settings (user_id, exchange_id, currency, chain, address, min_amount, created_at)
	VALUES ($1, $2, $3, $4, $5, $6, now())
	RETURNING *;

