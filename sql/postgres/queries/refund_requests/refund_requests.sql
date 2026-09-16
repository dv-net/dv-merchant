-- name: GetAllByUserIDAndStatus :many
SELECT rr.*,
       t.id AS transaction_id,
       t.tx_hash,
       t.amount,
       t.amount_usd,
       t.currency_id,
       c.code AS currency_code,
       t.blockchain,
       t.from_address,
       t.to_address,
       bt.risk_level,
       bt.score
FROM refund_requests rr
         INNER JOIN stores s ON rr.store_id = s.id
         INNER JOIN blocked_transactions bt ON bt.id = rr.blocked_transaction_id
         INNER JOIN transactions t ON t.id = bt.transaction_id
         INNER JOIN currencies c ON c.id = t.currency_id
WHERE s.user_id = $1 AND rr.status = $2
ORDER BY rr.created_at DESC;

-- name: GetByBlockedTransactionID :one
SELECT *
FROM refund_requests
WHERE blocked_transaction_id = $1
LIMIT 1;
