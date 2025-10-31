DROP INDEX IF EXISTS receipts_account_id_index;

ALTER TABLE receipts
RENAME COLUMN account_id TO wallet_id;

ALTER TABLE receipts
ADD CONSTRAINT invoices_payer_id_foreign
FOREIGN KEY (wallet_id) REFERENCES wallets(id);