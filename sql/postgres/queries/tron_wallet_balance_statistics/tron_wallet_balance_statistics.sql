-- name: InsertTronWalletBalanceStatisticsBatch :batchexec
INSERT INTO tron_wallet_balance_statistics (processing_owner_id,
                                            address,
                                            staked_bandwidth,
                                            staked_energy,
                                            delegated_energy,
                                            delegated_bandwidth,
                                            available_energy,
                                            available_bandwidth)
VALUES (sqlc.arg(processing_owner_id)::uuid,
        sqlc.arg(address)::varchar,
        sqlc.arg(staked_bandwidth)::numeric,
        sqlc.arg(staked_energy)::numeric,
        sqlc.arg(delegated_energy)::numeric,
        sqlc.arg(delegated_bandwidth)::numeric,
        sqlc.arg(available_energy)::numeric,
        sqlc.arg(available_bandwidth)::numeric)
ON CONFLICT (processing_owner_id, address, created_at) DO NOTHING;

-- name: ApproximateByResolution :many
SELECT
    processing_owner_id,
    AVG(staked_bandwidth)::numeric AS staked_bandwidth,
    AVG(staked_energy)::numeric AS staked_energy,
    AVG(delegated_energy)::numeric AS delegated_energy,
    AVG(delegated_bandwidth)::numeric AS delegated_bandwidth,
    AVG(available_bandwidth)::numeric AS available_bandwidth,
    AVG(available_energy)::numeric AS available_energy,
    DATE_TRUNC(sqlc.arg('resolution')::varchar, created_at AT TIME ZONE sqlc.arg('timezone')::varchar)::timestamp AS day
FROM tron_wallet_balance_statistics
WHERE created_at >= (sqlc.arg(date_from)::timestamp AT TIME ZONE sqlc.arg('timezone')::varchar)
  AND created_at <= (sqlc.arg(date_to)::timestamp AT TIME ZONE sqlc.arg('timezone')::varchar)
  AND (sqlc.arg(processing_owner)::uuid IS NULL OR processing_owner_id = sqlc.arg(processing_owner)::uuid)
GROUP BY processing_owner_id, DATE_TRUNC(sqlc.arg('resolution')::varchar, created_at AT TIME ZONE sqlc.arg('timezone')::varchar)
ORDER BY day, processing_owner_id;