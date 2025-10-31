DROP INDEX IF EXISTS unconfirmed_transactions_invoice_id_foreign;
ALTER TABLE unconfirmed_transactions DROP COLUMN IF EXISTS invoice_id;
ALTER TABLE unconfirmed_transactions RENAME COLUMN account_id TO wallet_id;
