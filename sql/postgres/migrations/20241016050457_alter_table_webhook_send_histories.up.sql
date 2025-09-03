ALTER TABLE webhook_send_histories
    ADD COLUMN IF NOT EXISTS store_id uuid;

CREATE INDEX IF NOT EXISTS idx_store_id ON webhook_send_histories (store_id);

UPDATE webhook_send_histories
SET store_id = (SELECT t.store_id FROM transactions AS t WHERE t.id = webhook_send_histories.tx_id)
WHERE type = 'PaymentReceived';