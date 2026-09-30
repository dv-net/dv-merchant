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

-- name: IsIPAllowed :one
SELECT (
    NOT EXISTS (
        SELECT 1 FROM store_whitelist sw
        WHERE sw.store_id = sqlc.arg(store_id)
    )
    OR EXISTS (
        SELECT 1 FROM store_whitelist sw
        WHERE sw.store_id = sqlc.arg(store_id)
          AND sw.ip::inet = sqlc.arg(ip)::inet
    )
)::bool;