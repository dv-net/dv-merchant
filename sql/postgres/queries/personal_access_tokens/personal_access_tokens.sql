-- name: GetByToken :one
SELECT * FROM personal_access_tokens WHERE (expires_at > now() OR expires_at IS NULL)  AND token=$1 LIMIT 1;

-- name: ClearAllByUser :exec
DELETE FROM personal_access_tokens WHERE (tokenable_type = 'user' and tokenable_id = $1 and name = 'AuthToken' and token != $2);