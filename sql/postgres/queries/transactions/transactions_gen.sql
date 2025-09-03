-- name: Create :one
INSERT INTO transactions (user_id, store_id, receipt_id, wallet_id, currency_id, blockchain, tx_hash, bc_uniq_key, type, from_address, to_address, amount, amount_usd, fee, withdrawal_is_manual, network_created_at, created_at, created_at_index, is_system)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, now(), extract(epoch from now()) * 1000, $17)
	RETURNING *;

-- name: GetById :one
SELECT * FROM transactions WHERE id=$1 LIMIT 1;

