-- name: GetPendingNotifications :many
SELECT * FROM notification_send_history
    WHERE sent_at IS NULL
    AND attempts < 2
    ORDER BY created_at
    LIMIT 500;

-- name: ExistWasSentRecently :one
SELECT EXISTS (
    SELECT 1
    FROM notification_send_history
    WHERE destination = $1
      AND type = $2
      AND channel = $3
      AND created_at >= now() - interval '30 minutes'
) AS was_sent_recently;
