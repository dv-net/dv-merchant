-- name: GetAllByUserIDAndStatus :many
SELECT rr.*
FROM refund_requests rr
INNER JOIN stores s ON rr.store_id = s.id
WHERE s.user_id = $1
  AND rr.status = $2
ORDER BY rr.created_at DESC;

-- name: GetByBlockedTransactionID :one
SELECT *
FROM refund_requests
WHERE blocked_transaction_id = $1
LIMIT 1;
