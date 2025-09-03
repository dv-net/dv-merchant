-- name: GetByEmail :one
SELECT *
FROM users
WHERE email = $1
LIMIT 1;

-- name: GetByID :one
SELECT *
FROM users
WHERE id = $1
LIMIT 1;

-- name: ChangePassword :one
UPDATE users
SET password=$1,
    updated_at=now()
WHERE id = $2
RETURNING *;

-- name: UpdateProcessingOwnerId :one
UPDATE users
SET processing_owner_id=$1,
    updated_at=now()
WHERE id = $2
RETURNING *;

-- name: UpdateEmailVerifiedAt :one 
UPDATE users
SET email_verified_at=now(),
    updated_at=now()
WHERE id = $1
RETURNING *;

-- name: UpdateBanned :one
UPDATE users
SET banned=$1,
    updated_at=now()
WHERE id = $2
RETURNING *;

-- name: UpdateExchange :one
UPDATE users
SET exchange_slug=$1,
    updated_at=now()
WHERE id = $2
RETURNING *;

-- name: GetAllWithExchangeEnabled :many
SELECT *
FROM users
WHERE exchange_slug IS NOT NULL;

-- name: UpdateRate :exec
UPDATE users
SET rate_scale=$1,
    rate_source = $2,
    updated_at=now()
WHERE id = $3;

-- name: ChangeEmail :exec
UPDATE users
SET email             = sqlc.arg(new_email),
    email_verified_at = null
WHERE email = sqlc.arg(old_email);

-- name: SetEmail :exec
UPDATE users
SET email             = $2,
    email_verified_at = null
WHERE id = $1;

-- name: UpdateDvToken :exec
UPDATE users
set dvnet_token = $2
WHERE id = $1;

-- name: GetUnverifed :many
SELECT *
FROM users
WHERE email_verified_at IS NULL
  AND created_at < NOW() - INTERVAL '30 day';

-- name: GetActiveProcessingOwnersWithTronDelegate :many
SELECT DISTINCT processing_owner_id
FROM users u
         INNER JOIN settings s
                    ON s.model_id = u.id and s.model_type = 'User' AND s.name = sqlc.arg(tron_setting_name) AND
                       s.value = sqlc.arg(tron_setting_value)
WHERE (u.banned IS NULL OR u.banned = false)
  AND u.deleted_at IS NULL
  AND u.processing_owner_id IS NOT NULL;