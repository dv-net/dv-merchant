-- name: Create :one
INSERT INTO tron_wallet_balance_statistics (processing_owner_id, address, staked_bandwidth, staked_energy, delegated_energy, delegated_bandwidth, available_bandwidth, available_energy, created_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, now())
	RETURNING *;

