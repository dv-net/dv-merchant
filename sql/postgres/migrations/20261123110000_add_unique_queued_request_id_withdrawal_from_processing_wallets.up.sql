-- Withdrawal idempotency by request_id: at most one queued (not yet sent to processing) withdrawal
-- per request_id. The service-level existence check alone raced under concurrent requests.
-- A retry after a failed transfer stays possible: such rows already have transfer_id set.
CREATE UNIQUE INDEX IF NOT EXISTS withdrawal_from_processing_wallets_request_id_queued_uniq
    ON withdrawal_from_processing_wallets (request_id)
    WHERE transfer_id IS NULL;
