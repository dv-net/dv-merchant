-- name: Create :one
INSERT INTO refund_requests (blocked_transaction_id, wallet_id, store_id, transfer_id, destination_address, status, email, reviewed_at, created_at, execute_after, last_error)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, now(), $9, $10)
	RETURNING *;

-- name: GetAllByWalletID :many
SELECT * FROM refund_requests WHERE wallet_id=$1;

-- name: GetById :one
SELECT * FROM refund_requests WHERE id=$1 LIMIT 1;

-- name: Update :one
UPDATE refund_requests
	SET transfer_id=$1, status=$2, reviewed_at=$3, updated_at=now(), execute_after=$4, last_error=$5
WHERE id=$6
	RETURNING *;
