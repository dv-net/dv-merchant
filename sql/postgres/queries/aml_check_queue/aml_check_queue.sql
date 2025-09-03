-- name: FetchPending :many
SELECT sqlc.embed(u),
       sqlc.embed(ac),
       sqlc.embed(acq),
       sqlc.embed(amls),
       acq.attempts >= sqlc.arg(max_queue_attempts) as is_last_attempt
FROM aml_check_queue acq
         INNER JOIN aml_checks ac ON acq.aml_check_id = ac.id AND ac.status = sqlc.arg(pending_status)
         INNER JOIN users u on ac.user_id = u.id AND (u.banned  IS NULL OR u.banned = false)
         INNER JOIN aml_services amls ON ac.service_id = amls.id
ORDER BY acq.created_at, acq.updated_at;

-- name: Delete :exec
DELETE
FROM aml_check_queue
WHERE id = $1;

-- name: IncrementAttempts :exec
UPDATE aml_check_queue SET attempts = attempts + 1 WHERE id = $1;
