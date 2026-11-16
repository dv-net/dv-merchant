UPDATE refund_requests SET auth_method = 'none' WHERE auth_method = 'api';
UPDATE refund_requests SET auth_method = 'email' WHERE auth_method = 'api_email';

ALTER TABLE refund_requests
    DROP CONSTRAINT IF EXISTS refund_requests_auth_method_check,
    ADD CONSTRAINT refund_requests_auth_method_check CHECK (auth_method IN ('email', 'none'));
