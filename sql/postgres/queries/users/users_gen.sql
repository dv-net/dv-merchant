-- name: Create :one
INSERT INTO users (email, email_verified_at, password, remember_token, processing_owner_id, location, language, rate_source, created_at, deleted_at, banned, exchange_slug, rate_scale, dvnet_token, two_fa_reset_expires_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, now(), $9, $10, $11, $12, $13, $14)
	RETURNING *;

-- name: Delete :exec
DELETE FROM users WHERE id=$1;

-- name: GetAll :many
SELECT * FROM users WHERE deleted_at IS NULL ORDER BY created_at DESC LIMIT $1 OFFSET $2;

-- name: Update :one
UPDATE users
	SET location=$1, language=$2, rate_source=$3, updated_at=now(), exchange_slug=$4
WHERE id=$5
	RETURNING *;
