-- name: GetByProcessID :one
SELECT * FROM logs WHERE process_id=$1 ORDER BY created_at DESC LIMIT 1;

-- name: DeleteLogBySlug :exec
DELETE FROM logs WHERE log_type_slug = $1 and process_id != $2;

-- name: GetLogsBySlug :many
WITH grouped_logs AS (
    SELECT
        process_id,
        JSONB_AGG(
                JSONB_BUILD_OBJECT(
                        'status', status,
                        'message', message,
                        'created_at', created_at
                ) ORDER BY created_at
        ) AS messages,
        MIN(created_at) AS created_at,
        MAX(CASE WHEN status = 'failed' THEN 1 ELSE 0 END) > 0 AS failure
    FROM logs
    WHERE log_type_slug = $1
    GROUP BY process_id
    LIMIT 100
)
SELECT
    process_id,
    failure,
    created_at,
    messages AS messages
FROM grouped_logs;
