-- name: GetQueuedWithCurrency :many
SELECT sqlc.embed(ubq), sqlc.embed(c)
FROM update_balance_queue ubq
         INNER JOIN currencies c ON ubq.currency_id = c.id
ORDER BY ubq.created_at;