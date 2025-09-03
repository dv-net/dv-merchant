-- name: GetByUserAndOrderID :one
SELECT exchange_withdrawal_history.*, exchanges.slug
FROM exchange_withdrawal_history
         LEFT JOIN exchanges ON exchange_withdrawal_history.exchange_id = exchanges.id
WHERE user_id = $1
  AND exchange_withdrawal_history.id = $2
  AND exchange_id = $3;

-- name: GetByID :one
SELECT *
FROM exchange_withdrawal_history
WHERE id = $1;

-- name: Update :exec
UPDATE exchange_withdrawal_history
SET updated_at        = NOW(),
    status            = COALESCE(sqlc.narg(status)::varchar, status),
    exchange_order_id = COALESCE(sqlc.narg(exchange_order_id), exchange_order_id),
    txid              = COALESCE(sqlc.narg(txid), txid),
    native_amount     = COALESCE(sqlc.narg(native_amount), native_amount),
    fiat_amount       = COALESCE(sqlc.narg(fiat_amount), fiat_amount),
    fail_reason       = COALESCE(sqlc.narg(fail_reason), fail_reason),
    exchange_connection_hash = COALESCE(sqlc.narg(exchange_connection_hash), exchange_connection_hash)
WHERE id = $1;
