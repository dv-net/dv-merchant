-- name: GetAllActiveWithUserKeys :many
SELECT e.slug, e.name, euk.value, ek.name as key_name, ek.title as key_title, euk.created_at as exchange_connected_at FROM exchanges e
         LEFT JOIN exchange_keys ek on e.id = ek.exchange_id
         LEFT JOIN exchange_user_keys euk on ek.id = euk.exchange_key_id AND euk.user_id = sqlc.arg(user_id)
         WHERE is_active=true;

-- name: GetExchangeKeysBySlug :many
SELECT ek.name, euk.id as user_key_id, ek.id as exchange_key_id, euk.value as existing_key_value
FROM exchanges e
         INNER JOIN exchange_keys ek ON e.id = ek.exchange_id
         LEFT JOIN exchange_user_keys euk ON euk.exchange_key_id = ek.id AND euk.user_id = sqlc.arg(user_id)
WHERE e.slug = sqlc.arg(slug);

-- name: GetExchangeBySlug :one
SELECT * FROM exchanges WHERE slug = sqlc.arg(slug);