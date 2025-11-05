-- name: Create :one
INSERT INTO unconfirmed_transactions (user_id, store_id, account_id, currency_id, tx_hash, bc_uniq_key, type, from_address, to_address, amount, amount_usd, network_created_at, created_at, blockchain, invoice_id)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, now(), $13, $14)
	RETURNING *;

-- name: GetById :one
SELECT * FROM unconfirmed_transactions WHERE id=$1 LIMIT 1;

