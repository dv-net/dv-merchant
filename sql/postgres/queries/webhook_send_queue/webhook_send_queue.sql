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
       (select count(distinct id)
        from webhook_send_histories
        where webhook_send_histories.status = 'failed'
          and webhook_send_histories.send_queue_job_id = whsq.id) as retries_count
FROM webhook_send_queue whsq
         join store_webhooks sw on whsq.webhook_id = sw.id and sw.enabled = true
         left join webhook_send_histories whsh on whsh.send_queue_job_id = whsq.id and whsh.status = 'success'
WHERE whsh.id is null
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