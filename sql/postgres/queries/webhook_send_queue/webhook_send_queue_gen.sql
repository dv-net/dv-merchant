-- name: Create :one
INSERT INTO webhook_send_queue (webhook_id, seconds_delay, transaction_id, payload, signature, event, last_sent_at, created_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, now())
	RETURNING *;

-- name: GetById :one
SELECT * FROM webhook_send_queue WHERE id=$1 LIMIT 1;

