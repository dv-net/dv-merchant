-- name: GetByStore :one
SELECT *
FROM wallets
WHERE deleted_at IS NULL
  AND store_id = $1
  AND store_external_id = $2
LIMIT 1;

-- name: GetFullDataByID :one
select sqlc.embed(w), sqlc.embed(st)
from wallets w
         left join stores st on w.store_id = st.id
where w.id = $1;

-- name: GetWalletWithStore :many
SELECT w.id as wallet_id, w.store_external_id, wa.address, wa.currency_id, sqlc.embed(s)
FROM wallets w
         INNER JOIN stores s ON w.store_id = s.id
         INNER JOIN wallet_addresses wa ON w.id = wa.wallet_id
WHERE wa.address = $1
  AND s.user_id = $2
  AND w.deleted_at IS NULL;

-- name: UpdateUserEmail :exec
UPDATE wallets
SET email=$1,
    updated_at = now()
WHERE id = $2;

-- name: UpdateUserUntrustedEmail :exec
UPDATE wallets
SET untrusted_email=$1,
    updated_at = now()
WHERE id = $2;

-- name: UpdateUserIPAddress :exec
UPDATE wallets
SET ip_address=$1,
    updated_at = now()
WHERE id = $2;

-- name: UpdateUserLocale :exec
UPDATE wallets
SET locale=$1,
    updated_at = now()
WHERE id = $2;

-- name: SearchByParam :many
WITH deposit_stats AS (
    SELECT t.to_address,
           t.currency_id,
           COUNT(*)      AS tx_count,
           SUM(t.amount) AS total_deposit
    FROM transactions t
    WHERE t.type = 'deposit'
    GROUP BY t.to_address, t.currency_id
)
SELECT w.id                                   AS "wallet_id",
       w.created_at                           AS "wallet_created_at",
       w.store_external_id                    AS "store_external_id",
       s.id                                   AS "store_id",
       s.name                                 AS "store_name",
       w.email                                AS "email",
       w.untrusted_email                      AS "untrusted_email",
       wa.address                             AS "address",
       wa.id                                  AS "wallet_address_id",
       wa.blockchain                          AS "blockchain",
       wa.currency_id                         AS "currency_id",
       wa.amount                              AS "amount",
       c.code                                 AS "currency_code",
       COALESCE(ds.tx_count, 0)::numeric      AS "deposits_count",
       COALESCE(ds.total_deposit, 0)::numeric AS "deposits_sum"
FROM wallets AS w
         JOIN wallet_addresses AS wa ON w.id = wa.wallet_id
         JOIN currencies AS c ON c.id = wa.currency_id
         JOIN stores AS s ON s.id = w.store_id
         JOIN user_stores us ON s.id = us.store_id AND us.user_id = wa.user_id
         LEFT JOIN deposit_stats ds ON ds.to_address = wa.address AND ds.currency_id = wa.currency_id
WHERE wa.user_id = sqlc.arg(user_id)
  AND (
    w.ip_address = sqlc.arg(criteria)
        OR w.email = sqlc.arg(criteria)
        OR w.untrusted_email = sqlc.arg(criteria)
        OR wa.address = sqlc.arg(criteria)
        OR wa.address = lower(sqlc.arg(criteria))
        OR w.store_external_id = sqlc.arg(criteria)
);

-- name: SoftDeleteByStore :many
UPDATE wallets SET deleted_at = now(), updated_at = now() WHERE store_id = $1 RETURNING id;

-- name: RestoreByStore :many
UPDATE wallets SET deleted_at = NULL, updated_at = now() WHERE store_id = $1 RETURNING id;