-- name: GetQueuedWithdrawalsWithCurrencyAndUser :many
SELECT sqlc.embed(wfpw), sqlc.embed(u), sqlc.embed(c)
FROM withdrawal_from_processing_wallets wfpw
         INNER JOIN currencies c on c.id = wfpw.currency_id
         INNER JOIN stores s on s.id = wfpw.store_id
         INNER JOIN users u on s.user_id = u.id
WHERE wfpw.transfer_id IS NULL;

-- name: GetPrefetchHistoryByUserID :many
SELECT sqlc.embed(wfpw), sqlc.embed(c), (wfpw.amount * rate.exchange_rate) ::decimal as amount_usd
FROM withdrawal_from_processing_wallets wfpw
         INNER JOIN currencies c on c.id = wfpw.currency_id
         INNER JOIN stores s on s.id = wfpw.store_id
         LEFT JOIN (SELECT unnest(sqlc.arg('currency_ids')::text[])     AS currency_id,
                           unnest(sqlc.arg('currency_rate')::decimal[]) AS exchange_rate) rate
                   on c.id = rate.currency_id
WHERE s.user_id = $1
  AND wfpw.transfer_id IS NULL;

-- name: UpdateTransferID :exec
UPDATE withdrawal_from_processing_wallets
SET transfer_id = sqlc.arg(transfer_id)::uuid,
    amount_usd  = sqlc.arg(amount_usd)::numeric
WHERE id = sqlc.arg(id)::uuid;

-- name: GetWithdrawalWithTransfer :one
SELECT sqlc.embed(wfpw),
       coalesce(t.kind, ''),
       coalesce(t.stage, ''),
       coalesce(t.status, ''),
       coalesce(t.tx_hash, ''),
       coalesce(t.message, '')
FROM withdrawal_from_processing_wallets wfpw
         LEFT JOIN transfers t ON wfpw.transfer_id = t.id
WHERE wfpw.id = $1
  AND wfpw.store_id = $2;

-- name: FindByTransferID :one
SELECT sqlc.embed(wfpw),
       coalesce(t.kind, ''),
       coalesce(t.stage, ''),
       coalesce(t.status, '')
FROM withdrawal_from_processing_wallets wfpw
         INNER JOIN transfers t ON wfpw.transfer_id = t.id
WHERE t.id = $1
LIMIT 1;

-- name: IsWithdrawalExistByRequestID :one
SELECT EXISTS (SELECT 1
               FROM withdrawal_from_processing_wallets wfpw
                        LEFT JOIN transfers t ON wfpw.transfer_id = t.id
               WHERE wfpw.request_id = $1
                 AND (
                   wfpw.transfer_id IS NULL
                       OR t.stage != 'failed'
                   ));

-- name: GetByID :one
SELECT *
FROM withdrawal_from_processing_wallets
WHERE id = $1
  AND store_id = $2;

-- name: DeleteWithdrawalFromProcessingWallets :one
DELETE
FROM withdrawal_from_processing_wallets
WHERE id = $1
  AND store_id = $2
  AND transfer_id IS NULL
RETURNING *;
