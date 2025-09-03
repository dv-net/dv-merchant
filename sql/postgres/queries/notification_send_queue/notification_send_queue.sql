-- name: GetQueuedNotifications :many
SELECT *
FROM notification_send_queue
WHERE attempts < sqlc.arg(max_attempts)::integer
ORDER BY created_at
LIMIT 500;

-- name: IncreaseAttempts :one
UPDATE notification_send_queue
SET attempts = attempts + 1
WHERE id = $1 RETURNING attempts;