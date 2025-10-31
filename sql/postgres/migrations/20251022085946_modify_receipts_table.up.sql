ALTER TABLE receipts
DROP CONSTRAINT IF EXISTS invoices_payer_id_foreign;

ALTER TABLE receipts
RENAME COLUMN wallet_id TO account_id;

CREATE INDEX IF NOT EXISTS receipts_account_id_index ON receipts(account_id);
