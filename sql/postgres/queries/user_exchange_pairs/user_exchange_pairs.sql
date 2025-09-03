-- name: Delete :exec
DELETE FROM user_exchange_pairs WHERE exchange_id=$1 AND user_id=$2;

-- name: UpdatePairs :batchexec
INSERT INTO user_exchange_pairs (exchange_id, user_id, currency_from, currency_to, symbol, type)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT DO NOTHING;

-- name: GetAll :many
SELECT * FROM user_exchange_pairs;

-- name: DeleteByUserAndExchangeID :exec
DELETE FROM user_exchange_pairs WHERE user_id = $1 AND exchange_id = $2;