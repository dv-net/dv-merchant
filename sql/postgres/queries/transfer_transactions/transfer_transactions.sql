-- name: BatchCreate :batchexec
INSERT INTO transfer_transactions (transfer_id, tx_hash, bandwidth_amount, energy_amount, native_token_amount,
                                   native_token_fee, tx_type, status, step, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, now(), now())
ON CONFLICT (transfer_id, tx_type, tx_hash) DO UPDATE SET bandwidth_amount    = $3,
                                                          energy_amount       = $4,
                                                          native_token_amount = $5,
                                                          native_token_fee    = $6,
                                                          status              = $8,
                                                          step                = $9,
                                                          updated_at          = now();

-- name: CalculateTransfersExpense :many
SELECT COUNT(DISTINCT t.id)                                                                 AS transfers_count,
       SUM(tt.native_token_fee)::numeric                                                    AS total_trx_fee,
       SUM(tt.bandwidth_amount)::numeric                                                    AS total_bandwidth,
       SUM(tt.energy_amount)::numeric                                                       AS total_energy,
       DATE_TRUNC(sqlc.arg('resolution')::varchar, tt.created_at AT TIME ZONE sqlc.arg('timezone')::varchar)::timestamp AS day
FROM transfer_transactions tt
         INNER JOIN transfers t ON tt.transfer_id = t.id AND t.user_id = sqlc.arg('user_id')::uuid
         INNER JOIN currencies c ON t.currency_id = c.id AND c.blockchain = sqlc.arg('blockchain')
    AND tt.created_at >= (sqlc.arg('date_from')::timestamp AT TIME ZONE sqlc.arg('timezone')::varchar)
    AND tt.created_at < (sqlc.arg('date_to')::timestamp AT TIME ZONE sqlc.arg('timezone')::varchar)
    AND tt.tx_type = ANY (sqlc.arg('tx_types')::varchar[])
GROUP BY DATE_TRUNC(sqlc.arg('resolution')::varchar, tt.created_at AT TIME ZONE sqlc.arg('timezone')::varchar)
ORDER BY day;