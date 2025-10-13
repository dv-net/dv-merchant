-- name: GetTransactionsByStoreId :many
SELECT *
FROM transactions
WHERE store_id = $1
ORDER BY created_at_index DESC
LIMIT $2 OFFSET $3;

-- name: GetTransactionsByUserId :many
SELECT *
FROM transactions
WHERE user_id = $1
ORDER BY created_at_index DESC
LIMIT $2 OFFSET $3;

-- name: GetTransactionsByStoreAndType :many
SELECT *
FROM transactions
WHERE store_id = $1
  and type = $2
ORDER BY created_at_index DESC
LIMIT $3 OFFSET $4;

-- name: GetLastByHashAndBlockchain :one
SELECT *
FROM transactions
WHERE tx_hash = $1
  AND blockchain = $2
ORDER BY created_at_index DESC
LIMIT 1;

-- name: GetExistingWithdrawalAddress :one
SELECT coalesce(to_address, '')::varchar
FROM transactions
WHERE from_address = sqlc.arg(from_addr)::varchar
  AND to_address = ANY (sqlc.arg(withdraw_addresses)::varchar[])
  AND type = 'transfer'
LIMIT 1;

-- name: GetTransactionByHashAndBcUniqKey :one
SELECT *
FROM transactions
WHERE tx_hash = $1
  and bc_uniq_key = $2
LIMIT 1;

-- name: GetAddressBalance :one
SELECT SUM(
               CASE
                   WHEN type = 'deposit' AND to_address = $1 THEN amount
                   WHEN type = 'transfer' AND from_address = $1 THEN -amount
                   ELSE 0
                   END)::numeric AS balance
FROM transactions
WHERE $1 IN (to_address, from_address)
  and currency_id = $2
LIMIT 1;

-- name: GetBalanceNativeToken :one
SELECT (SUM(CASE
                WHEN type = 'deposit' AND transactions.to_address = $1 THEN amount
                WHEN type = 'transfer' AND transactions.from_address = $1 THEN -amount
                ELSE 0
    END)
    - COALESCE((SELECT SUM(fee)
                FROM transactions
                WHERE type = 'transfer'
                  AND transactions.from_address = $1
                  AND transactions.blockchain = $3), 0))::numeric AS balance

FROM transactions
WHERE $1 IN (to_address, from_address)
  AND transactions.currency_id = $2
  AND transactions.blockchain = $3
LIMIT 1;

-- name: CalculateDepositStatistics :many
WITH tx_filtered AS (SELECT id,
                            amount_usd,
                            currency_id,
                            date_trunc(sqlc.arg(resolution)::text, (created_at AT TIME ZONE 'UTC') AT TIME ZONE
                                                                   sqlc.arg(timezone)::text) AS truncated_date
                     FROM transactions
                     WHERE type = 'deposit'
                       AND amount_usd >= 1
                       AND (created_at >=
                            (sqlc.arg(date_from)::timestamp AT TIME ZONE sqlc.arg(timezone)::text AT TIME ZONE 'UTC')
                         AND created_at <
                             (sqlc.arg(date_to)::timestamp AT TIME ZONE sqlc.arg(timezone)::text AT TIME ZONE 'UTC'))
                       AND (sqlc.arg(store_uuids)::uuid[] IS NULL OR store_id = ANY (sqlc.arg(store_uuids)::uuid[]))
                       AND (sqlc.arg(currency_ids)::varchar[] IS NULL OR
                            currency_id = ANY (sqlc.arg(currency_ids)::varchar[]))
                       AND user_id = sqlc.arg(user_id)::uuid
                       AND is_system = false),
     currency_stats AS (SELECT truncated_date,
                               currency_id,
                               COUNT(id)       AS currency_tx_count,
                               SUM(amount_usd) AS currency_sum_usd
                        FROM tx_filtered
                        GROUP BY truncated_date, currency_id),
     monthly_totals AS (SELECT truncated_date,
                               COUNT(id)       AS total_tx_count,
                               SUM(amount_usd) AS total_sum_usd
                        FROM tx_filtered
                        GROUP BY truncated_date)
SELECT mt.truncated_date::timestamp  AS date,
       mt.total_tx_count             AS tx_count,
       mt.total_sum_usd::numeric     AS sum,
       sqlc.arg(resolution)::varchar AS resolution,
       jsonb_object_agg(
               cs.currency_id,
               jsonb_build_object('tx_count', cs.currency_tx_count, 'sum_usd', cs.currency_sum_usd::numeric)
       )                             AS currency_stats
FROM monthly_totals mt
         LEFT JOIN currency_stats cs ON mt.truncated_date = cs.truncated_date
GROUP BY mt.truncated_date, mt.total_tx_count, mt.total_sum_usd
ORDER BY mt.truncated_date DESC;

-- name: CalculateDepositStatisticsTotal :one
WITH currency_stats AS (SELECT currency_id,
                               COUNT(id)       AS currency_tx_count,
                               SUM(amount_usd) AS currency_sum_usd
                        FROM transactions
                        WHERE type = 'deposit'
                          AND amount_usd >= 1
                          AND created_at >=
                              (sqlc.arg(date_from)::timestamp AT TIME ZONE sqlc.arg(timezone)::text AT TIME ZONE 'UTC')
                          AND created_at <
                              (sqlc.arg(date_to)::timestamp AT TIME ZONE sqlc.arg(timezone)::text AT TIME ZONE 'UTC')
                          AND (sqlc.arg(store_uuids)::uuid[] IS NULL OR store_id = ANY (sqlc.arg(store_uuids)::uuid[]))
                          AND user_id = sqlc.arg(user_id)::uuid
                          AND is_system = false
                        GROUP BY currency_id),
     overall_stats AS (SELECT COUNT(id)       AS tx_count,
                              SUM(amount_usd) AS sum
                       FROM transactions
                       WHERE type = 'deposit'
                         AND amount_usd >= 1
                         AND created_at >=
                             (sqlc.arg(date_from)::timestamp AT TIME ZONE sqlc.arg(timezone)::text AT TIME ZONE 'UTC')
                         AND created_at <
                             (sqlc.arg(date_to)::timestamp AT TIME ZONE sqlc.arg(timezone)::text AT TIME ZONE 'UTC')
                         AND (sqlc.arg(store_uuids)::uuid[] IS NULL OR store_id = ANY (sqlc.arg(store_uuids)::uuid[]))
                         AND user_id = sqlc.arg(user_id)::uuid
                         AND is_system = false)
SELECT os.tx_count     as tx_count,
       os.sum::numeric as sum,
       jsonb_object_agg(
               cs.currency_id,
               jsonb_build_object(
                       'tx_count', cs.currency_tx_count,
                       'sum_usd', cs.currency_sum_usd::numeric
               )
       )               AS currency_stats
FROM overall_stats os,
     currency_stats cs
GROUP BY os.tx_count, os.sum;


-- name: FindTransactionByHashAndUserID :one
WITH tx AS ((SELECT t.id,
                    true         as is_confirmed,
                    t.user_id,
                    t.store_id,
                    t.receipt_id as receipt_id,
                    t.wallet_id,
                    t.currency_id,
                    t.bc_uniq_key,
                    t.blockchain as blockchain,
                    t.tx_hash,
                    t.from_address,
                    t.to_address,
                    t.amount,
                    t.amount_usd,
                    t.fee        as fee,
                    t.type,
                    t.network_created_at,
                    t.created_at
             FROM transactions t
             where t.tx_hash = $1
               AND t.user_id = $2)
            UNION
            (SELECT ut.id,
                    false as is_confirmed,
                    ut.user_id,
                    ut.store_id,
                    null  as receipt_id,
                    ut.wallet_id,
                    ut.currency_id,
                    ut.bc_uniq_key,
                    ut.blockchain,
                    ut.tx_hash,
                    ut.from_address,
                    ut.to_address,
                    ut.amount,
                    ut.amount_usd,
                    0     as fee,
                    ut.type,
                    ut.network_created_at,
                    ut.created_at
             FROM unconfirmed_transactions ut
             where ut.tx_hash = $1
               AND ut.user_id = $2)
            LIMIT 1)
SELECT tx.*,
       w.id                as wallet_id,
       w.store_id          as wallet_store_id,
       w.store_external_id as store_external_id,
       w.created_at        as wallet_created_at,
       w.updated_at        as wallet_updated_at,
       c.code              as currency_code,
       c.blockchain        as currency_blockchain
FROM tx
         INNER JOIN wallets w on w.id = tx.wallet_id
         INNER JOIN currencies c on c.id = tx.currency_id
LIMIT 1;

-- name: FindLastWalletTransactions :many
WITH tx AS ((SELECT true as is_confirmed,
                    t.wallet_id,
                    t.currency_id,
                    t.tx_hash,
                    t.amount,
                    t.amount_usd,
                    t.type,
                    t.created_at
             FROM transactions t
             WHERE t.wallet_id = $1
               AND t.amount_usd >= 1
               AND t.type = $2
             ORDER BY created_at_index DESC
             LIMIT $3)
            UNION
            (SELECT false as is_confirmed,
                    ut.wallet_id,
                    ut.currency_id,
                    ut.tx_hash,
                    ut.amount,
                    ut.amount_usd,
                    ut.type,
                    ut.created_at
             FROM unconfirmed_transactions ut
             WHERE ut.wallet_id = $1
               AND ut.type = $2
               AND ut.amount_usd >= 1
               AND NOT EXISTS (SELECT 1
                               FROM transactions t2
                               WHERE t2.tx_hash = ut.tx_hash
                                 AND t2.currency_id = ut.currency_id
                                 AND t2.bc_uniq_key = ut.bc_uniq_key)
             ORDER BY created_at DESC
             LIMIT $3)
            LIMIT $3)
SELECT tx.*,
       c.id as curr_code
FROM tx
         INNER JOIN currencies c on c.id = tx.currency_id
LIMIT $3;

-- name: GetWalletTransactions :many
SELECT *
FROM transactions t
WHERE t.wallet_id = $1
  AND t.to_address = $2
ORDER BY network_created_at DESC
LIMIT 500;