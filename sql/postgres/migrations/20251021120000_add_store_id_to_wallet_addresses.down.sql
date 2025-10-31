-- Drop index
DROP INDEX IF EXISTS idx_wallet_addresses_store_id;

-- Drop foreign key constraint
ALTER TABLE wallet_addresses DROP CONSTRAINT IF EXISTS wallet_addresses_store_id_foreign;

-- Drop store_id column
ALTER TABLE wallet_addresses DROP COLUMN IF EXISTS store_id;
