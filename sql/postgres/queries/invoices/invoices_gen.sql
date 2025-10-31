-- name: Create :one
INSERT INTO invoices (user_id, store_id, order_id, expected_amount_usd, received_amount_usd, status, expires_at, created_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, now())
	RETURNING *;

