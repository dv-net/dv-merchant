
ALTER TABLE transactions RENAME COLUMN wallet_id TO account_id;

ALTER TABLE transactions
ADD COLUMN invoice_id uuid NULL
CONSTRAINT transactions_invoice_id_foreign
REFERENCES invoices(id);


CREATE INDEX transactions_invoice_id_index ON public.transactions (invoice_id);