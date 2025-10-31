-- name: GetByID :one
SELECT *
FROM invoices
WHERE id = $1;

-- name: UpdateStatus :one
UPDATE invoices
SET status = $2, updated_at = now()
WHERE id = $1
    RETURNING *;

-- name: UpdateReceivedAmount :one
UPDATE invoices
SET received_amount_usd = $2, updated_at = now()
WHERE id = $1
    RETURNING *;

-- name: GetExpiredInvoices :many
SELECT *
FROM invoices
WHERE expires_at < NOW()
  AND status IN ('pending', 'underpaid')
ORDER BY expires_at ASC;

-- name: ExpireInvoices :exec
UPDATE invoices
SET status = 'expired', updated_at = NOW()
WHERE expires_at < NOW()
  AND status IN ('pending');
