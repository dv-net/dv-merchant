-- name: GetByUserId :many
SELECT r.* from receipts r inner join stores s on r.store_id = s.id where s.user_id = $1 limit $2 offset $3;