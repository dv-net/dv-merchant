-- name: GetByUserID :many
SELECT * FROM user_address_book 
WHERE user_id = $1 AND deleted_at IS NULL
ORDER BY submitted_at DESC;

-- name: GetByID :one
SELECT * FROM user_address_book 
WHERE id = $1 AND deleted_at IS NULL;

-- name: GetByUserAndAddress :one
SELECT * FROM user_address_book 
WHERE user_id = $1 AND address = $2 AND currency_id = $3 AND type = $4;

-- name: GetByUserAndAddressAllCurrencies :many
SELECT * FROM user_address_book 
WHERE user_id = $1 AND address = $2 AND type = $3 AND deleted_at IS NULL;

-- name: GetByUserAddressAndBlockchain :many
SELECT * FROM user_address_book 
WHERE user_id = $1 AND address = $2 AND blockchain = $3 AND type = $4 AND deleted_at IS NULL;

-- name: GetByUserAndCurrency :many
SELECT * FROM user_address_book 
WHERE user_id = $1 AND currency_id = $2 AND deleted_at IS NULL
ORDER BY submitted_at DESC;

-- name: GetByUserAndBlockchain :many
SELECT * FROM user_address_book 
WHERE user_id = $1 AND blockchain = $2 AND deleted_at IS NULL
ORDER BY submitted_at DESC;

-- name: Create :one
INSERT INTO user_address_book (user_id, address, currency_id, name, tag, blockchain, type, submitted_at, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
RETURNING *;

-- name: Update :one
UPDATE user_address_book 
SET name = $2, tag = $3, updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: SoftDelete :exec
UPDATE user_address_book 
SET deleted_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
WHERE id = $1;

-- name: Delete :exec
DELETE FROM user_address_book 
WHERE id = $1;

-- name: CheckExists :one
SELECT EXISTS(
    SELECT 1 FROM user_address_book 
    WHERE user_id = $1 AND address = $2 AND currency_id = $3 AND type = $4
) AS exists;

-- name: CheckExistsWithTrashed :one
SELECT EXISTS(
    SELECT 1 FROM user_address_book 
    WHERE user_id = $1 AND address = $2 AND currency_id = $3
) AS exists;

-- name: GetTrashedEntry :one
SELECT * FROM user_address_book 
WHERE user_id = $1 AND address = $2 AND currency_id = $3 AND deleted_at IS NOT NULL;

-- name: RestoreFromTrash :one
UPDATE user_address_book 
SET deleted_at = NULL, name = $5, tag = $6, updated_at = CURRENT_TIMESTAMP
WHERE user_id = $1 AND address = $2 AND currency_id = $3 AND type = $4 AND deleted_at IS NOT NULL
RETURNING *;