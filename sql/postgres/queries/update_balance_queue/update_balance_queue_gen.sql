-- name: Create :one
INSERT INTO update_balance_queue (currency_id, address, native_token_balance_update)
	VALUES ($1, $2, $3)
	RETURNING *;

-- name: Delete :exec
DELETE FROM update_balance_queue WHERE id=$1;

