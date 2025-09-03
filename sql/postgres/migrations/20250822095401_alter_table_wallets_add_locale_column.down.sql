-- Remove locale column from wallets table
ALTER TABLE wallets DROP COLUMN IF EXISTS locale;