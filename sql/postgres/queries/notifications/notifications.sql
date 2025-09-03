-- name: GetAllTypes :many
SELECT * FROM notifications WHERE category IS NULL;
