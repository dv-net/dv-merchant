-- name: Create :exec
INSERT INTO notification_send_history (destination, message_text, sender, created_at, type, channel, notification_send_queue_id, store_id, user_id)
	VALUES ($1, $2, $3, now(), $4, $5, $6, $7, $8);

-- name: GetById :one
SELECT * FROM notification_send_history WHERE id=$1 LIMIT 1;

-- name: Update :one
UPDATE notification_send_history
	SET updated_at=now(), sent_at=$1, notification_send_queue_id=$2, store_id=$3, user_id=$4
	WHERE destination=$5 AND id=$6
	RETURNING *;

