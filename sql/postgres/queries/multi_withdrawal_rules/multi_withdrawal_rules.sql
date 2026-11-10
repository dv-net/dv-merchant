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

-- name: DisableIfManualAddressMissing :exec
UPDATE multi_withdrawal_rules mwr
SET mode           = 'disabled',
    manual_address = null,
    updated_at     = now()
WHERE mwr.withdrawal_wallet_id = sqlc.arg(withdrawal_wallet_id)::uuid
  AND mwr.manual_address IS NOT NULL
  AND NOT EXISTS (SELECT 1
                  FROM withdrawal_wallet_addresses wwa
                  WHERE wwa.withdrawal_wallet_id = mwr.withdrawal_wallet_id
                    AND wwa.address = mwr.manual_address
                    AND wwa.deleted_at IS NULL);
