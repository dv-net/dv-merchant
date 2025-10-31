BEGIN;

ALTER TABLE wallet_addresses
    ADD COLUMN IF NOT EXISTS status VARCHAR(32) DEFAULT  "static",
    ADD COLUMN IF NOT EXISTS account_type VARCHAR(32),
    ADD COLUMN IF NOT EXISTS account_id UUID;

UPDATE wallet_addresses
SET account_type = 'wallet',
    account_id   = wallet_id
WHERE wallet_id IS NOT NULL
  AND account_type IS NULL;

DO $$
    BEGIN
        IF EXISTS (
            SELECT 1 FROM wallet_addresses WHERE account_type IS NULL
        ) THEN
            RAISE EXCEPTION 'Some rows have NULL owner_type. Migration aborted.';
        END IF;
    END$$;

ALTER TABLE wallet_addresses
    ALTER COLUMN account_type SET NOT NULL,
    ALTER COLUMN account_type SET DEFAULT 'wallet';

CREATE INDEX IF NOT EXISTS idx_wallet_addresses_owner
    ON wallet_addresses (account_type, account_id);


ALTER TABLE wallet_addresses DROP COLUMN wallet_id;

COMMIT;
