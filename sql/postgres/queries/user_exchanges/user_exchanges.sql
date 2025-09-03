-- name: GetByUserID :one
SELECT * FROM user_exchanges WHERE user_id = $1 AND exchange_id = $2;

-- name: GetByUserAndExchangeID :one
SELECT * FROM user_exchanges WHERE user_id = $1 AND exchange_id = $2;

-- name: CreateByUserID :exec
WITH active_user_exchanges AS (
    SELECT DISTINCT u.id AS user_id, e.id AS exchange_id
    FROM users u
             LEFT JOIN exchange_user_keys euk ON euk.user_id = u.id
             LEFT JOIN exchange_keys ek ON ek.id = euk.exchange_key_id
             LEFT JOIN exchanges e ON e.id = ek.exchange_id
    WHERE euk.exchange_key_id IS NOT NULL
    AND u.id = $1
)
INSERT INTO user_exchanges (user_id, exchange_id, withdrawal_state, swap_state)
SELECT user_id, exchange_id, 'disabled', 'disabled'
FROM active_user_exchanges
ON CONFLICT (user_id, exchange_id) DO NOTHING;

-- name: DeleteByKeyID :exec
WITH exchange_key AS (
    SELECT DISTINCT euk.user_id, ek.exchange_id FROM exchange_user_keys euk
    LEFT JOIN exchange_keys ek ON ek.id = euk.exchange_key_id
    WHERE euk.id = $1
)
DELETE FROM user_exchanges ue
USING exchange_key
WHERE ue.exchange_id = exchange_key.exchange_id
AND ue.user_id = exchange_key.user_id;

-- name: DeleteByUserAndExchangeID :exec
DELETE FROM user_exchanges WHERE user_id = $1 AND exchange_id = $2;