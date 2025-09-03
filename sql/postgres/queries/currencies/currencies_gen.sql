-- name: Create :one
INSERT INTO currencies (id, code, name, precision, is_fiat, blockchain, contract_address, withdrawal_min_balance, has_balance, status, sort_order, min_confirmation, created_at, is_stablecoin, currency_label, token_label)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
	RETURNING *;

-- name: GetAll :many
SELECT * FROM currencies;

-- name: GetByID :one
SELECT * FROM currencies WHERE id=$1 LIMIT 1;

