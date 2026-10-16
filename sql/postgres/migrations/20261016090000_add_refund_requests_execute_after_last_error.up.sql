ALTER TABLE refund_requests
    ADD COLUMN IF NOT EXISTS execute_after timestamp,
    ADD COLUMN IF NOT EXISTS last_error    text;

-- refund requests are updated by transfer status webhooks
CREATE INDEX IF NOT EXISTS refund_requests_transfer_id_idx ON refund_requests (transfer_id);

-- refund transfers keep the link to their request, also after the request is detached from a failed transfer
ALTER TABLE transfers
    ADD COLUMN IF NOT EXISTS refund_request_id uuid REFERENCES refund_requests (id);
