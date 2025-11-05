ALTER TABLE unconfirmed_transactions RENAME COLUMN wallet_id TO account_id;

ALTER TABLE unconfirmed_transactions
    ADD COLUMN invoice_id uuid NULL
        CONSTRAINT unconfirmed_transactions_invoice_id_foreign
            REFERENCES invoices(id);


CREATE INDEX unconfirmed_transactionss_invoice_id_index ON unconfirmed_transactions (invoice_id);