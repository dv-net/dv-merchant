ALTER TABLE transfers
    DROP COLUMN IF EXISTS refund_request_id;

DROP INDEX IF EXISTS refund_requests_transfer_id_idx;

ALTER TABLE refund_requests
    DROP COLUMN IF EXISTS last_error,
    DROP COLUMN IF EXISTS execute_after;
