-- name: CheckStoreHasUser :one
SELECT EXISTS(SELECT * FROM user_stores WHERE store_id = $1 and user_id = $2);

-- name: GetStoreIDsByUser :many
SELECT distinct(store_id) FROM user_stores WHERE user_id = sqlc.arg(user_id)
AND (sqlc.arg(store_uuids)::uuid[] IS NULL OR store_id = ANY (sqlc.arg(store_uuids)::uuid[]));