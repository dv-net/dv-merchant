DROP INDEX IF EXISTS transactions_invoice_id_index;
ALTER TABLE transactions DROP COLUMN IF EXISTS invoice_id;
ALTER TABLE transactions RENAME COLUMN account_id TO wallet_id;
