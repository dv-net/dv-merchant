-- name: Delete :exec
DELETE FROM user_exchange_pairs WHERE exchange_id=$1 AND user_id=$2;

-- name: UpdatePairs :batchexec
INSERT INTO user_exchange_pairs (exchange_id, user_id, currency_from, currency_to, symbol, type)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT DO NOTHING;

-- name: GetAll :many
SELECT uep.* FROM user_exchange_pairs uep
INNER JOIN user_exchanges ue ON ue.user_id = uep.user_id AND ue.exchange_id = uep.exchange_id
WHERE ue.swap_state = 'enabled';

-- name: DeleteByUserAndExchangeID :exec
DELETE FROM user_exchange_pairs WHERE user_id = $1 AND exchange_id = $2;