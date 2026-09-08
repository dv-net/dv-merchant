-- name: GetQueuedWebhooks :many
SELECT whsq.id,
       whsq.webhook_id,
       whsq.seconds_delay,
       whsq.transaction_id,
       whsq.event,
       whsq.payload,
       whsq.signature,
       whsq.created_at,
       whsq.last_sent_at,
       sw.store_id,
       sw.url,
       count(whsh_failed.id) as retries_count
FROM webhook_send_queue whsq
         join store_webhooks sw on whsq.webhook_id = sw.id and sw.enabled = true
         left join webhook_send_histories whsh_success
                   on whsh_success.send_queue_job_id = whsq.id and whsh_success.status = 'success'
         left join webhook_send_histories whsh_failed
                   on whsh_failed.send_queue_job_id = whsq.id and whsh_failed.status = 'failed'
WHERE whsh_success.id is null
  AND (whsq.last_sent_at is null
       OR whsq.last_sent_at + make_interval(secs => whsq.seconds_delay) <= now())
GROUP BY whsq.id, sw.id
ORDER BY whsq.created_at
LIMIT 500;

-- name: UpdateDelay :exec
UPDATE webhook_send_queue
set seconds_delay=sqlc.arg(delay), last_sent_at = now()
where id = $1;

-- name: Delete :exec
DELETE
FROM webhook_send_queue
WHERE id = $1;