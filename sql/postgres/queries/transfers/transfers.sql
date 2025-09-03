-- name: IsFailedTransferExistsByAddress :one
SELECT EXISTS(SELECT 1
              FROM transfers
              where user_id = $1
                AND kind = 'transferFromAddress'
                AND $2 = ANY (from_addresses)
                AND currency_id = $3
                AND created_at > (NOW() - INTERVAL '20 minutes'));

-- name: UpdateTransferStatus :exec
UPDATE transfers
SET status = $1,
    stage = $2,
    message = $3,
    step = $4,
    updated_at = now()
WHERE id = $5;

-- name: UpdateTxHash :exec
UPDATE transfers SET tx_hash = $2 WHERE id = $1 AND tx_hash is null;

-- name: BatchDeleteTransfers :batchexec
DELETE FROM transfers WHERE id = $1 AND user_id = $2 AND status = 'failed';