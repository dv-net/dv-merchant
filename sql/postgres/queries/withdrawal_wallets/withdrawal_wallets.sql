-- name: GetWalletsForWithdrawal :many
SELECT *
FROM withdrawal_wallets
WHERE blockchain = $1
  and withdrawal_enabled = 'enabled';

-- name: GetWithdrawalWallets :many
SELECT *
FROM withdrawal_wallets
WHERE withdrawal_wallets.user_id = sqlc.arg(user_id)::uuid;

-- name: GetWithdrawalWalletByCurrency :one
SELECT *
FROM withdrawal_wallets
WHERE withdrawal_wallets.user_id = $1
  AND currency_id = $2;

-- name: Update :one
UPDATE withdrawal_wallets
SET withdrawal_enabled=$1,
    withdrawal_min_balance=$2,
    withdrawal_min_balance_usd=$3,
    withdrawal_interval=$4,
    updated_at=now()
WHERE currency_id = $5
  AND user_id = $6
RETURNING *;

-- name: GetForMultiWithdrawal :many
SELECT sqlc.embed(u),
       sqlc.embed(mwr),
       sqlc.embed(curr),
       sqlc.embed(ww),
       array_agg(wwa.address) FILTER (WHERE wwa.address IS NOT NULL)::varchar[] AS addresses
FROM withdrawal_wallets ww
         INNER JOIN multi_withdrawal_rules mwr
                    ON ww.id = mwr.withdrawal_wallet_id
                        AND mwr.mode != 'disabled'
         INNER JOIN users u
                    ON ww.user_id = u.id
                        AND (u.banned = false OR u.banned IS NULL)
         INNER JOIN currencies curr
                    ON ww.currency_id = curr.id
         LEFT JOIN withdrawal_wallet_addresses wwa
                   ON ww.id = wwa.withdrawal_wallet_id AND wwa.deleted_at IS NULL
WHERE ww.blockchain = ANY (ARRAY ['bitcoin', 'litecoin', 'bitcoincash', 'dogecoin']::varchar[])
  AND (sqlc.narg(user_id)::uuid IS NULL OR u.id = sqlc.narg(user_id)::uuid)
GROUP BY u.id, curr.id, mwr.id, ww.id;