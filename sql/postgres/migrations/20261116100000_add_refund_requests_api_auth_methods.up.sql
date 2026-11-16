-- refunds requested by the store through the external api:
-- api - to the deposit sender without a confirmation code, api_email - confirmed by the code sent to the wallet email
ALTER TABLE refund_requests
    DROP CONSTRAINT IF EXISTS refund_requests_auth_method_check,
    ADD CONSTRAINT refund_requests_auth_method_check CHECK (auth_method IN ('email', 'none', 'api', 'api_email'));
