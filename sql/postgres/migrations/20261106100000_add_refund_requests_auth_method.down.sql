ALTER TABLE refund_requests
    DROP COLUMN auth_method;

-- requests created without the email verification have no email, a placeholder keeps them
UPDATE refund_requests SET email = 'unverified' WHERE email IS NULL;

ALTER TABLE refund_requests
    ALTER COLUMN email SET NOT NULL,
    DROP CONSTRAINT IF EXISTS refund_requests_email_check,
    ADD CONSTRAINT refund_requests_email_check CHECK (email != '');
