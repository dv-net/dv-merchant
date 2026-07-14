ALTER TABLE stores
    DROP COLUMN verification_status,
    DROP COLUMN verified_at,
    DROP COLUMN verified_by,
    DROP COLUMN rejection_reason;
