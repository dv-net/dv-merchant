-- name: BatchInsertExchangeAddresses :batchexec
INSERT INTO exchange_addresses (exchange_id, address, chain, currency, address_type, user_id, create_type)
VALUES ($1, $2, $3, $4, $5, $6, $7)
ON CONFLICT DO NOTHING;

-- name: Update :one
UPDATE exchange_addresses
SET address=$1
WHERE id=$2 AND user_id=$3
RETURNING *;

-- name: DeleteByUser :exec
DELETE FROM exchange_addresses WHERE exchange_id=$1 AND user_id=$2;

-- name: DeleteByUserAndExchangeID :exec
DELETE FROM exchange_addresses WHERE exchange_id=$1 AND user_id=$2;

-- name: GetAllDepositAddress :many
SELECT ea.address,
       ea.currency,
       ea.chain,
       ea.address_type,
       MAX(exchanges.slug)::VARCHAR as slug,
       MAX(exchanges.name)::VARCHAR as name,
       MAX(ec.ticker)::VARCHAR      as ticker
FROM exchange_addresses AS ea
         LEFT JOIN exchanges ON ea.exchange_id = exchanges.id
         LEFT JOIN exchange_chains ec ON ea.chain = ec.chain AND ea.currency = ec.currency_id
WHERE ea.user_id = $1
  AND ea.address_type = 'deposit'
GROUP BY ea.address,
         ea.currency,
         ea.chain,
         ea.address_type;