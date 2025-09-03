-- name: CheckWebhookWasSent :one
SELECT EXISTS(SELECT 1 from webhook_send_histories where tx_id=$1 and type=$2 and url=$3 and status='success');

-- name: GetFailedAttemptsCount :one
SELECT COUNT(id) FROM webhook_send_histories WHERE tx_id=$1 AND type=$2 AND url=$3 AND status='failed';

-- name: GetAllByTxID :many
SELECT * FROM webhook_send_histories where tx_id=$1 ORDER BY created_at DESC;