-- name: Create :one
INSERT INTO notification_send_queue (destination, type, parameters, channel, created_at, args)
	VALUES ($1, $2, $3, $4, now(), $5)
	RETURNING *;

-- name: Delete :exec
DELETE FROM notification_send_queue WHERE id=$1;

-- name: GetById :one
SELECT * FROM notification_send_queue WHERE id=$1 LIMIT 1;

