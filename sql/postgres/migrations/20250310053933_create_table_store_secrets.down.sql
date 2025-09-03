ALTER TABLE IF EXISTS store_webhooks
    ADD COLUMN IF NOT EXISTS secret varchar(255);

UPDATE store_webhooks sw
SET secret = ss.secret
FROM store_secrets ss
WHERE sw.store_id = ss.store_id;

ALTER TABLE store_webhooks
    ALTER COLUMN secret SET NOT NULL;

DROP TABLE IF EXISTS store_secrets;