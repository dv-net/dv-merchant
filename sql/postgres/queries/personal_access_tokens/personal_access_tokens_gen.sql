-- name: Create :one
INSERT INTO personal_access_tokens (tokenable_type, tokenable_id, name, token, expires_at, created_at)
	VALUES ($1, $2, $3, $4, $5, $6)
	RETURNING *;

-- name: Delete :exec
DELETE FROM personal_access_tokens WHERE id=$1;

