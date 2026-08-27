-- name: GetUnclaimedByWalletID :many
SELECT bt.*
FROM blocked_transactions bt
WHERE bt.wallet_id = $1
  AND NOT EXISTS (
    SELECT 1
    FROM refund_requests rr
    WHERE rr.blocked_transaction_id = bt.id
  )
ORDER BY bt.created_at DESC;
