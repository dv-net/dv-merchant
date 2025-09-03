-- name: Delete :exec
DELETE FROM store_whitelist WHERE store_id=$1;

-- name: DeleteByIP :exec
DELETE FROM store_whitelist WHERE ip=$1 AND store_id=$2;

-- name: CheckExistsByIP :one
SELECT (
    EXISTS (
        SELECT * FROM store_whitelist WHERE ip = $1 AND store_id = $2
    )
);