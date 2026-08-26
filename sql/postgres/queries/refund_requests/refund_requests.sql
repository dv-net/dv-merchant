-- name: GetAllByStoreIDAndStatus :many
SELECT *
FROM refund_requests
WHERE store_id = $1
  AND status = $2
ORDER BY created_at DESC;

-- name: GetByBlockedTransactionID :one
SELECT *
FROM refund_requests
WHERE blocked_transaction_id = $1
LIMIT 1;
