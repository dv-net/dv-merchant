ALTER TABLE stores
    ADD COLUMN verification_status varchar(255) NOT NULL DEFAULT 'PENDING',
    ADD COLUMN verified_at timestamp NULL,
    ADD COLUMN verified_by uuid NULL REFERENCES users(id),
    ADD COLUMN rejection_reason text NULL;
