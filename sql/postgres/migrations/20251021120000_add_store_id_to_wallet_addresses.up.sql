-- Add store_id column to wallet_addresses
ALTER TABLE wallet_addresses
ADD COLUMN store_id uuid NULL;

-- Fill store_id from wallets table where account_type = 'wallet'
UPDATE wallet_addresses wa
SET store_id = w.store_id
FROM wallets w
WHERE wa.account_type = 'wallet'
  AND wa.account_id = w.id;

-- Add foreign key constraint
ALTER TABLE wallet_addresses
ADD CONSTRAINT wallet_addresses_store_id_foreign
FOREIGN KEY (store_id) REFERENCES stores(id);

-- Add index for store_id
CREATE INDEX idx_wallet_addresses_store_id ON wallet_addresses(store_id);