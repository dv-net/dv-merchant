-- name: Create :one
INSERT INTO wallets (store_id, store_external_id, created_at, email, ip_address, untrusted_email, locale)
	VALUES ($1, $2, now(), $3, $4, $5, $6)
	RETURNING *;

-- name: GetById :one
SELECT * FROM wallets WHERE deleted_at IS NULL AND id=$1 LIMIT 1;

