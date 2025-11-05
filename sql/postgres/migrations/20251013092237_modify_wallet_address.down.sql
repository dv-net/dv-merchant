BEGIN;
ALTER TABLE wallet_addresses
    ADD COLUMN IF NOT EXISTS wallet_id UUID;

UPDATE wallet_addresses
SET wallet_id = account_id
WHERE account_type = 'wallet';

ALTER TABLE public.wallet_addresses
    DROP COLUMN IF EXISTS account_type,
    DROP COLUMN IF EXISTS account_id;

DROP INDEX IF EXISTS idx_wallet_addresses_owner;
COMMIT;