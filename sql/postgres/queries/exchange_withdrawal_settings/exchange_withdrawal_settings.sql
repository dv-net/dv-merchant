-- name: GetExisting :one
SELECT * FROM exchange_withdrawal_settings
    WHERE user_id = $1 AND exchange_id = $2 AND currency = $3 AND chain = $4;

-- name: ChangeMinAmount :exec
UPDATE exchange_withdrawal_settings
    SET min_amount = $1
    WHERE id = $2 AND user_id = $3;

-- name: ChangeAddress :exec
UPDATE exchange_withdrawal_settings
SET address = $1
WHERE id = $2 AND user_id = $3;

-- name: GetAllByUser :many
SELECT * FROM exchange_withdrawal_settings
    WHERE user_id = $1 AND exchange_id = $2;

-- name: GetByID :one
SELECT * FROM exchange_withdrawal_settings
WHERE user_id = $1 AND exchange_id = $2 AND id = $3;

-- name: Delete :exec
DELETE FROM exchange_withdrawal_settings WHERE id = $1 AND user_id = $2 AND exchange_id = $3;

-- name: GetActive :many
SELECT * FROM exchange_withdrawal_settings where is_enabled = true;

-- name: DeleteByUserAndExchangeID :exec
DELETE FROM exchange_withdrawal_settings WHERE user_id = $1 AND exchange_id = $2;

-- name: Update :one
UPDATE exchange_withdrawal_settings
SET min_amount = COALESCE(sqlc.narg(min_amount), min_amount),
    is_enabled = COALESCE(sqlc.narg(enabled), is_enabled)
WHERE id = sqlc.arg(id)
  AND user_id = sqlc.arg(user_id)
RETURNING *;

-- name: GetAllAddressesWithEnabledCurr :many
SELECT distinct ews.address, c.blockchain
FROM exchange_withdrawal_settings ews
         INNER JOIN currencies c on c.id = ews.currency AND c.status = true
WHERE user_id = $1;

-- name: UpdateIsEnabledByID :one
UPDATE exchange_withdrawal_settings
SET is_enabled = sqlc.arg(new_state)
WHERE id = sqlc.arg(id) AND user_id = sqlc.arg(user_id)
RETURNING *;

-- name: UpdateIsEnabledByUserAndExchangeID :exec
UPDATE exchange_withdrawal_settings
SET is_enabled = $3
WHERE user_id = $2 AND exchange_id = $1;