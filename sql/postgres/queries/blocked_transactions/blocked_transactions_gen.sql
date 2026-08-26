-- name: Create :one
INSERT INTO blocked_transactions (user_id, store_id, transaction_id, aml_check_id, wallet_id, risk_level, score, created_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, now())
	RETURNING *;

-- name: GetAllByWalletID :many
SELECT * FROM blocked_transactions WHERE wallet_id=$1;

-- name: GetById :one
SELECT * FROM blocked_transactions WHERE id=$1 LIMIT 1;
