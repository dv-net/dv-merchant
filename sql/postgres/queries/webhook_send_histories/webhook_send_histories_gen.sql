-- name: Create :one
INSERT INTO webhook_send_histories (tx_id, send_queue_job_id, type, url, status, request, response, response_status_code, created_at, is_manual, store_id)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, now(), $9, $10)
	RETURNING *;

-- name: GetById :one
SELECT * FROM webhook_send_histories WHERE id=$1 LIMIT 1;

