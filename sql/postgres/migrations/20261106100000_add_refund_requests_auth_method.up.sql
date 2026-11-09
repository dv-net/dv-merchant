-- a refund to the deposit sender may be requested without the email verification (user setting refund_auth_mode),
-- then the request has no email
ALTER TABLE refund_requests
    ALTER COLUMN email DROP NOT NULL,
    DROP CONSTRAINT IF EXISTS refund_requests_email_check,
    ADD CONSTRAINT refund_requests_email_check CHECK (email IS NULL OR email != '');

ALTER TABLE refund_requests
    ADD COLUMN auth_method varchar(32) NOT NULL DEFAULT 'email' CHECK (auth_method IN ('email', 'none'));
