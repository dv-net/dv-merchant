-- name: GetAllWithTxByWalletID :many
SELECT bt.*,
       t.tx_hash,
       t.amount,
       t.amount_usd,
       t.currency_id,
       c.code AS currency_code,
       t.blockchain,
       t.from_address,
       t.to_address
FROM blocked_transactions bt
         INNER JOIN transactions t ON t.id = bt.transaction_id
         INNER JOIN currencies c ON c.id = t.currency_id
WHERE bt.wallet_id = $1
ORDER BY bt.created_at DESC;

-- name: GetUnclaimedByWalletID :many
SELECT bt.*,
       t.tx_hash,
       t.amount,
       t.amount_usd,
       t.currency_id,
       c.code AS currency_code,
       t.blockchain,
       t.from_address,
       t.to_address
FROM blocked_transactions bt
         INNER JOIN transactions t ON t.id = bt.transaction_id
         INNER JOIN currencies c ON c.id = t.currency_id
WHERE bt.wallet_id = $1
  AND NOT EXISTS (
    SELECT 1
    FROM refund_requests rr
    WHERE rr.blocked_transaction_id = bt.id
  )
ORDER BY bt.created_at DESC;
