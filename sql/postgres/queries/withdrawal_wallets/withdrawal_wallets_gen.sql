-- name: Create :one
INSERT INTO withdrawal_wallets (user_id, blockchain, currency_id, withdrawal_min_balance, withdrawal_interval, created_at, withdrawal_enabled, withdrawal_min_balance_usd)
	VALUES ($1, $2, $3, $4, $5, now(), $6, $7)
	RETURNING *;

-- name: GetById :one
SELECT * FROM withdrawal_wallets WHERE deleted_at IS NULL AND id=$1 LIMIT 1;

