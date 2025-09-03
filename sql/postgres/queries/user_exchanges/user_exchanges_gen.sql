-- name: Create :one
INSERT INTO user_exchanges (exchange_id, user_id, withdrawal_state, swap_state)
	VALUES ($1, $2, $3, $4)
	RETURNING *;

