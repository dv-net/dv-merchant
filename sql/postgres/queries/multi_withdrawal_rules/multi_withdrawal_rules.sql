-- name: CreateOrUpdate :exec
INSERT INTO multi_withdrawal_rules (withdrawal_wallet_id, mode, manual_address, created_at, updated_at)
VALUES ($1, $2, $3, now(), now())
ON CONFLICT (withdrawal_wallet_id) DO UPDATE SET mode           = $2,
                                                 manual_address = $3,
                                                 updated_at     = now();
-- name: GetByWalletID :one
SELECT *
FROM multi_withdrawal_rules
WHERE withdrawal_wallet_id = $1
LIMIT 1;

-- name: RemoveByWalletID :exec
DELETE
FROM multi_withdrawal_rules
WHERE withdrawal_wallet_id = $1;