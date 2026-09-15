-- name: GetAllByUserIDAndStatus :many
SELECT rr.*, t.amount, t.currency_id, t.tx_hash, t.blockchain
FROM refund_requests rr
         INNER JOIN stores s ON rr.store_id = s.id
         INNER JOIN blocked_transactions bt ON bt.id = rr.blocked_transaction_id
         INNER JOIN transactions t ON t.id = bt.transaction_id
WHERE s.user_id = $1 AND rr.status = $2
ORDER BY rr.created_at DESC;

-- name: GetByBlockedTransactionID :one
SELECT *
FROM refund_requests
WHERE blocked_transaction_id = $1
LIMIT 1;
