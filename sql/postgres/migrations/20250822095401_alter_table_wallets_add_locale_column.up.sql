-- Add locale column to wallets table
ALTER TABLE wallets ADD COLUMN locale VARCHAR(10) NOT NULL DEFAULT 'en';