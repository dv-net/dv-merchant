-- name: GetAddressesList :many
select address
from withdrawal_wallet_addresses
where withdrawal_wallet_id = $1::uuid
  and deleted_at is null;

-- name: GetAddresses :many
select *
from withdrawal_wallet_addresses
where withdrawal_wallet_id = $1::uuid
  and deleted_at is null;

-- name: GetByAddress :one
select *
from withdrawal_wallet_addresses
where withdrawal_wallet_id = sqlc.arg(withdrawal_wallet_id)::uuid
  and address = $1;

-- name: GetByAddressWithTrashed :one
select *
from withdrawal_wallet_addresses
where withdrawal_wallet_id = sqlc.arg(withdrawal_wallet_id)::uuid
  and address = $1;

-- name: UpdateDeletedAddress :one
update withdrawal_wallet_addresses
set name       = $1,
    deleted_at = NULL,
    updated_at = now()
where id = sqlc.arg(id)::uuid
RETURNING *;

-- name: SoftDelete :exec
update withdrawal_wallet_addresses
set deleted_at = now(),
    updated_at = now()
where id = sqlc.arg(id)::uuid;

-- name: SoftDeleteUnmatchedByAddress :exec
update withdrawal_wallet_addresses
set deleted_at = now(),
    updated_at = now()
where withdrawal_wallet_id = sqlc.arg(wallet_id)
  and (address != ANY (sqlc.arg('address')::varchar[]) or cardinality(sqlc.arg('address')::varchar[]) = 0);

-- name: SoftBatchDelete :exec
update withdrawal_wallet_addresses
set deleted_at = now(),
    updated_at = now()
where id = ANY (sqlc.arg('id')::uuid[]);

-- name: GetWithdrawalWalletsByBlockchain :many
select distinct (wa.address)
from withdrawal_wallet_addresses wa
         left join withdrawal_wallets w on w.id = wa.withdrawal_wallet_id
where w.blockchain = $1
  and w.user_id = $2
and wa.deleted_at is null;

-- name: GetWithdrawalAddressByCurrencyID :many
select distinct (wa.address)
from withdrawal_wallet_addresses wa
         left join withdrawal_wallets w on w.id = wa.withdrawal_wallet_id
where w.blockchain = $1
  and w.user_id = $2
  and w.currency_id = $3
  and wa.deleted_at is null;

-- name: UpdateList :batchexec
INSERT INTO withdrawal_wallet_addresses (address, name, withdrawal_wallet_id, created_at, updated_at)
VALUES ($1, $2, $3, now(), now())
ON CONFLICT (withdrawal_wallet_id, address) DO UPDATE
    SET name       = $2,
        updated_at = now(),
        deleted_at = null;

-- name: GetAddressWithCurrencyByUserID :many
SELECT distinct address,
       currencies.blockchain as blockchain
FROM withdrawal_wallet_addresses
         INNER JOIN
     withdrawal_wallets ON withdrawal_wallet_addresses.withdrawal_wallet_id = withdrawal_wallets.id
         INNER JOIN currencies ON withdrawal_wallets.currency_id = currencies.id AND currencies.status = true
WHERE user_id = $1;

-- name: CheckAddressExists :one
SELECT EXISTS (
    SELECT 1
    FROM withdrawal_wallet_addresses
    WHERE withdrawal_wallet_id = sqlc.arg(withdrawal_wallet_id)
      AND address = sqlc.arg(address)
      AND deleted_at IS NULL
) AS exists;