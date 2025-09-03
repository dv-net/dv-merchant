ALTER TABLE IF EXISTS unconfirmed_transactions ADD COLUMN IF NOT EXISTS blockchain VARCHAR(255);
UPDATE unconfirmed_transactions
SET blockchain = currencies.blockchain
    FROM currencies
WHERE unconfirmed_transactions.currency_id = currencies.id
  AND unconfirmed_transactions.blockchain IS NULL;
ALTER TABLE IF EXISTS unconfirmed_transactions ALTER COLUMN blockchain SET not null;
ALTER TABLE IF EXISTS unconfirmed_transactions DROP CONSTRAINT IF EXISTS unconfirmed_transactions_tx_id_bc_key_unique;
ALTER TABLE IF EXISTS unconfirmed_transactions ADD CONSTRAINT unconfirmed_transactions_tx_id_bc_key_unique
    unique (blockchain, tx_hash, bc_uniq_key);