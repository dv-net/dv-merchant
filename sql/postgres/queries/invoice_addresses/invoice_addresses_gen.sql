-- name: Create :one
INSERT INTO invoice_addresses (invoice_id, wallet_address_id, rate_at_creation, created_at)
	VALUES ($1, $2, $3, now())
	RETURNING *;

