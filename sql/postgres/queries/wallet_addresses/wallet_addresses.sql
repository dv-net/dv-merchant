-- name: GetWalletAddressesByWalletId :many
SELECT *
FROM wallet_addresses
WHERE deleted_at IS NULL
  AND wallet_id = $1;

-- name: GetWalletAddressesByAddress :one
SELECT *
FROM wallet_addresses
WHERE deleted_at IS NULL
  AND wallet_id = $1
  and address = $2
  and currency_id = $3
limit 1;

-- name: MarkAddressDirty :one
UPDATE wallet_addresses
SET updated_at=now(),
    dirty= true
WHERE address = $1
RETURNING *;

-- name: UpdateWalletBalance :exec
WITH balances as (SELECT SUM(
                                 CASE
                                     WHEN type = 'deposit' AND to_address = $1 THEN amount
                                     WHEN type = 'transfer' AND from_address = $1 THEN -amount
                                     ELSE 0
                                     END
                         )::numeric AS balance
                  FROM transactions
                  WHERE $1 IN (to_address, from_address)
                    and currency_id = $2
                  LIMIT 1)
UPDATE wallet_addresses wa
SET amount=COALESCE(b.balance, wa.amount),
    updated_at=now()
FROM balances b
WHERE wa.currency_id = $2
  AND wa.address = $1;

-- name: UpdateWalletNativeTokenBalance :exec
WITH native_token_balance as (SELECT (SUM(CASE
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
                              LIMIT 1)
UPDATE wallet_addresses wa
SET amount=COALESCE(b.balance, wa.amount),
    updated_at=now()
FROM native_token_balance b
WHERE wa.currency_id = $2
  AND wa.address = $1;

-- name: GetAllClearByWalletID :many
SELECT *
FROM wallet_addresses
WHERE wallet_id = $1
  AND dirty = false
  AND deleted_at IS NULL
  and currency_id = ANY (sqlc.arg('currency_ids')::varchar[]);

-- name: GetByWalletIDAndCurrencyID :one
SELECT *
FROM wallet_addresses
WHERE wallet_id = $1
  AND currency_id = $2
  AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT 1;

-- name: GetPrefetchWalletAddressByUserID :many
select withdrawal_wallets.id             as withdrawal_wallet_id,
       sqlc.embed(wallet_addresses),
       sqlc.embed(currencies),
       (amount * exchange_rate)::decimal as amount_usd
from wallet_addresses
         left join currencies
                   on wallet_addresses.currency_id = currencies.id
         left join
     (SELECT unnest(sqlc.arg('currency_ids')::text[])     AS currency_id,
             unnest(sqlc.arg('currency_rate')::decimal[]) AS exchange_rate) rate
     on currencies.id = rate.currency_id
         left join withdrawal_wallets on wallet_addresses.currency_id = withdrawal_wallets.currency_id
    and wallet_addresses.user_id = withdrawal_wallets.user_id
where wallet_addresses.user_id = $1
  and wallet_addresses.amount > withdrawal_wallets.withdrawal_min_balance
  and (withdrawal_wallets.withdrawal_min_balance_usd is null or
       (wallet_addresses.amount * exchange_rate)::decimal > withdrawal_wallets.withdrawal_min_balance_usd::numeric)
  and wallet_addresses.address not in (select unnest(from_addresses)
                                       from transfers
                                       where user_id = $1
                                         and kind = 'from_address'
                                         and (transfers.stage = 'in_progress'
                                           or
                                              (transfers.stage = 'completed' and
                                               updated_at > (now() - interval '5 minutes'))
                                           or
                                              (transfers.stage = 'failed' and
                                               created_at > (now() - interval '30 minutes'))
                                           ))
  and withdrawal_wallets.withdrawal_enabled = 'enabled'
order by amount_usd desc;

-- name: GetAddressForWithdrawal :one
select wallet_addresses.*, currencies.*, (amount * exchange_rate)::decimal as amount_usd
from wallet_addresses
         left join currencies
                   on wallet_addresses.currency_id = currencies.id
         left join
     (SELECT unnest(sqlc.arg('currency_ids')::text[])     AS currency_id,
             unnest(sqlc.arg('currency_rate')::decimal[]) AS exchange_rate) rate
     on currencies.id = rate.currency_id
where wallet_addresses.user_id = $1
  and wallet_addresses.currency_id = $2
  and wallet_addresses.amount >= $3
  and wallet_addresses.address not in (select unnest(from_addresses)
                                       from transfers
                                       where user_id = $1
                                         and kind = 'from_address'
                                         and transfers.currency_id = $2
                                         and transfers.blockchain = $4
                                         and (transfers.stage = 'in_progress'
                                           or
                                              (transfers.stage = 'completed' and
                                               updated_at > (now() - interval '5 minutes'))
                                           or
                                              (transfers.stage = 'failed' and
                                               created_at > (now() - interval '30 minutes'))
                                           ))
  and (sqlc.arg('min_usd')::numeric is null or (amount * exchange_rate)::decimal >= sqlc.arg('min_usd')::numeric)
order by amount_usd desc
limit 1;


-- name: GetWalletAddressesByUserID :many
SELECT *
FROM wallet_addresses
WHERE user_id = $1
  AND deleted_at IS NULL;

-- name: GetWalletAddressesTotalWithCurrencyID :many
SELECT amount::NUMERIC as balance, currencies.code
FROM wallet_addresses
         LEFT JOIN currencies ON wallet_addresses.currency_id = currencies.id
WHERE user_id = $1
  AND amount > 0
  AND deleted_at IS NULL;

-- name: GetHotWalletsTotalBalanceWithDust :one
SELECT
    COALESCE(SUM(wallet_addresses.amount * rate.exchange_rate), 0)::numeric as total_usd,
    COALESCE(SUM(CASE
        WHEN (wallet_addresses.amount * rate.exchange_rate) < 1
        THEN wallet_addresses.amount * rate.exchange_rate
        ELSE 0
    END), 0)::numeric as dust_usd
FROM wallet_addresses
LEFT JOIN (
    SELECT
        unnest(sqlc.arg('currency_ids')::text[]) AS currency_id,
        unnest(sqlc.arg('currency_rate')::decimal[]) AS exchange_rate
) rate ON wallet_addresses.currency_id = rate.currency_id
WHERE wallet_addresses.user_id = $1
  AND wallet_addresses.amount > 0
  AND wallet_addresses.deleted_at IS NULL;

-- name: FilterOwnerWalletAddresses :many
SELECT address, user_id, id as wallet_addresses_id
FROM wallet_addresses
WHERE user_id = $1
  AND address in (select unnest($2::text[]))
  AND deleted_at IS NULL;

-- name: GetWalletsDataForRestoreByBlockchains :many
SELECT sqlc.embed(wa), sqlc.embed(c), sqlc.embed(s)
FROM wallet_addresses wa
         JOIN currencies c ON wa.currency_id = c.id
         JOIN wallets w ON wa.wallet_id = w.id
         JOIN stores s ON w.store_id = s.id
WHERE (sqlc.arg(blockchains)::varchar[] IS NULL OR c.blockchain = ANY (sqlc.arg(blockchains)::varchar[]));

-- name: GetListByCurrencyWithAmount :one
select sqlc.embed(c), array_agg(address)::varchar[] as addresses, sum(amount)::numeric as amount
from wallet_addresses wa
         inner join currencies c on c.id = wa.currency_id
WHERE wa.currency_id = sqlc.arg(curr_id)
  AND wa.user_id = sqlc.arg(user_id)
  AND wa.amount > 0
  AND (COALESCE(array_length(sqlc.arg(ids)::uuid[], 1), 0) = 0 OR wa.id = ANY (sqlc.arg(ids)::uuid[]))
  AND (sqlc.arg(excluded_ids)::uuid[] IS NULL OR NOT wa.id = ANY (sqlc.arg(excluded_ids)::uuid[]))
group by c.id;

-- name: IsWalletExistsByAddress :one
SELECT EXISTS(SELECT * FROM wallet_addresses WHERE address = $1)
LIMIT 1;

-- name: SoftDeleteByWallets :exec
UPDATE wallet_addresses
SET deleted_at = now()
WHERE wallet_id = ANY ($1::uuid[]);

-- name: RestoreByWallets :exec
UPDATE wallet_addresses
SET deleted_at = NULL
WHERE wallet_id = ANY ($1::uuid[]);

-- name: GetAddressForMultiWithdrawal :one
SELECT wa.currency_id,
       array_agg(distinct wa.address)::varchar[]        AS addresses,
       SUM(wa.amount)::numeric                          AS total_amount,
       (SUM(wa.amount) * MAX(r.exchange_rate))::numeric AS amount_usd
FROM wallet_addresses wa
         INNER JOIN currencies c
                    ON wa.currency_id = c.id
         LEFT JOIN (SELECT unnest(sqlc.arg('currency_ids')::text[])     AS currency_id,
                           unnest(sqlc.arg('currency_rate')::decimal[]) AS exchange_rate) r
                   ON c.id = r.currency_id
WHERE wa.user_id = sqlc.arg(user_id)
  AND wa.currency_id = sqlc.arg(currency)
  AND wa.address NOT IN (SELECT unnest(from_addresses)
                         FROM transfers
                         WHERE user_id = sqlc.arg(user_id)
                           AND kind = 'from_address'
                           AND transfers.currency_id = sqlc.arg(currency)
                           AND transfers.blockchain = sqlc.arg(blockchain)
                           AND (
                             transfers.stage = 'in_progress'
                                 OR (
                                 transfers.stage = 'completed'
                                     AND updated_at > (NOW() - INTERVAL '5 minutes')
                                 )
                                 OR (
                                 transfers.stage = 'failed'
                                     AND created_at > (NOW() - INTERVAL '30 minutes')
                                 )
                             ))
  AND wa.amount > 0
  AND (
    ((sqlc.narg(min_usd)::numeric IS NULL OR sqlc.narg(min_usd)::numeric = 0) AND wa.amount > 0)
        -- Only addresses with low balance (lower than min for withdrawal) is required --
        OR (wa.amount * r.exchange_rate)::decimal < sqlc.narg(min_usd)::numeric
    )
  AND (
    ((sqlc.narg(min_amount)::numeric IS NULL OR sqlc.narg(min_amount) = 0) AND wa.amount > 0)
        -- Only addresses with low balance (lower than min for withdrawal) is required --
        OR wa.amount < sqlc.narg(min_amount)::numeric
    )
GROUP BY wa.user_id, wa.currency_id
HAVING (sqlc.narg(min_amount) IS NULL OR SUM(wa.amount) >= sqlc.narg(min_amount)::numeric)
   AND (sqlc.narg(min_usd) IS NULL OR (SUM(wa.amount) * MAX(r.exchange_rate))::numeric > sqlc.narg(min_usd))
ORDER BY (SUM(wa.amount) * MAX(r.exchange_rate))::decimal DESC
LIMIT 1;