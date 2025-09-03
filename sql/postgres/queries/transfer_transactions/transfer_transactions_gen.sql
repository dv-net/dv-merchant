-- name: Create :one
INSERT INTO transfer_transactions (transfer_id, tx_hash, bandwidth_amount, energy_amount, native_token_amount, native_token_fee, tx_type, status, step, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, now(), $10)
	RETURNING *;

