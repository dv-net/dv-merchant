-- name: Create :one
INSERT INTO multi_withdrawal_rules (withdrawal_wallet_id, mode, manual_address, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5)
	RETURNING *;

