-- name: Create :exec
INSERT INTO wallet_address_status_history (wallet_address_id, old_status, new_status)
	VALUES ($1, $2, $3);

