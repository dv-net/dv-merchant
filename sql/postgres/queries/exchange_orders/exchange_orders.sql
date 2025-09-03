-- name: Update :exec
UPDATE exchange_orders
SET updated_at        = now(),
    status            = COALESCE(sqlc.narg(status)::varchar, status),
    fail_reason       = COALESCE(sqlc.narg(fail_reason), fail_reason),
    amount            = COALESCE(sqlc.narg(amount), amount),
    amount_usd        = COALESCE(sqlc.narg(amount_usd), amount_usd),
    exchange_order_id = COALESCE(sqlc.narg(exchange_order_id), exchange_order_id),
    client_order_id   = COALESCE(sqlc.narg(client_order_id), client_order_id),
    exchange_connection_hash = COALESCE(sqlc.narg(exchange_connection_hash), exchange_connection_hash)
WHERE id = sqlc.arg(id);