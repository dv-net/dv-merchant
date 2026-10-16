-- name: GetAllByUserIDAndStatus :many
SELECT rr.*,
       t.id AS transaction_id,
       t.tx_hash,
       t.amount,
       t.amount_usd,
       t.currency_id,
       c.code AS currency_code,
       t.blockchain,
       t.from_address,
       t.to_address,
       bt.risk_level,
       bt.score
FROM refund_requests rr
         INNER JOIN stores s ON rr.store_id = s.id
         INNER JOIN blocked_transactions bt ON bt.id = rr.blocked_transaction_id
         INNER JOIN transactions t ON t.id = bt.transaction_id
         INNER JOIN currencies c ON c.id = t.currency_id
WHERE s.user_id = $1 AND rr.status = $2
ORDER BY rr.created_at DESC;

-- name: GetByBlockedTransactionID :one
SELECT *
FROM refund_requests
WHERE blocked_transaction_id = $1
LIMIT 1;

-- name: UpdateStatusIfCurrent :one
UPDATE refund_requests
	SET status=sqlc.arg(new_status), updated_at=now()
WHERE id=sqlc.arg(id) AND status=sqlc.arg(current_status)
	RETURNING *;

-- name: UpdateStatusByTransferID :execrows
-- Moves the refund request of the transfer to the new status when it is in one of the current statuses.
-- reset_transfer detaches the failed or canceled transfer, so the refund can be approved again.
UPDATE refund_requests
	SET status=sqlc.arg(new_status),
	    last_error=sqlc.narg(last_error),
	    transfer_id=CASE WHEN sqlc.arg(reset_transfer)::boolean THEN NULL ELSE transfer_id END,
	    execute_after=CASE WHEN sqlc.arg(reset_transfer)::boolean THEN NULL ELSE execute_after END,
	    updated_at=now()
WHERE transfer_id=sqlc.arg(transfer_id)
  AND status=ANY(sqlc.arg(current_statuses)::varchar[]);

-- name: SetLastError :exec
UPDATE refund_requests
	SET last_error=sqlc.narg(last_error), updated_at=now()
WHERE id=sqlc.arg(id);

-- name: ClaimForSending :one
-- Claims a pending review request for sending and attaches its refund transfer. It is committed
-- before the refund is sent to processing, so status webhooks always find the request.
UPDATE refund_requests
	SET status=sqlc.arg(new_status),
	    transfer_id=sqlc.arg(transfer_id),
	    reviewed_at=COALESCE(sqlc.narg(reviewed_at), reviewed_at),
	    execute_after=NULL,
	    updated_at=now()
WHERE id=sqlc.arg(id) AND status=sqlc.arg(current_status)
	RETURNING *;

-- name: MarkSent :one
-- Sets the result of the refund in processing. The request may be already moved by a status webhook,
-- then nothing is updated.
UPDATE refund_requests
	SET status=sqlc.arg(new_status),
	    execute_after=sqlc.narg(execute_after),
	    last_error=NULL,
	    updated_at=now()
WHERE id=sqlc.arg(id) AND transfer_id=sqlc.arg(transfer_id) AND status=sqlc.arg(current_status)
	RETURNING *;

-- name: RevertUnsent :execrows
-- Returns the request to pending review when processing did not create the refund transfer.
UPDATE refund_requests
	SET status=sqlc.arg(new_status),
	    transfer_id=NULL,
	    execute_after=NULL,
	    last_error=sqlc.narg(last_error),
	    updated_at=now()
WHERE id=sqlc.arg(id) AND transfer_id=sqlc.arg(transfer_id) AND status=sqlc.arg(current_status);

-- name: GetInFlightForReconcile :many
-- Requests sent to processing which were not updated for a while: a lost webhook or an unknown
-- result of the processing call.
SELECT *
FROM refund_requests
WHERE status = ANY(sqlc.arg(statuses)::varchar[])
  AND transfer_id IS NOT NULL
  AND updated_at < sqlc.arg(updated_before)::timestamp
ORDER BY updated_at
LIMIT sqlc.arg(max_items)::int;
