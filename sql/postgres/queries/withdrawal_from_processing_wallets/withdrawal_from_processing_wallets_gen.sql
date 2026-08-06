-- name: Create :one
INSERT INTO withdrawal_from_processing_wallets (currency_id, store_id, address_from, address_to, amount, amount_usd, request_id, blocked_by_processing_error)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	RETURNING *;
