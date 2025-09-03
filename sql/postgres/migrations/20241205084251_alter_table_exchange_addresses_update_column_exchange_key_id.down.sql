ALTER TABLE exchange_addresses ADD COLUMN IF NOT EXISTS exchange_key_id uuid NULL;
WITH exchange_ids AS (
    SELECT ea.address_type, ea.id, ek.id as exchange_key_id, ea.address FROM exchange_addresses ea
    LEFT JOIN exchange_keys ek ON ek.exchange_id = ea.exchange_id
)
UPDATE exchange_addresses SET exchange_key_id = exchange_ids.exchange_key_id
FROM exchange_ids
WHERE exchange_addresses.id = exchange_ids.id AND exchange_addresses.address = exchange_ids.address
  AND exchange_addresses.address_type = 'deposit';
ALTER TABLE exchange_addresses ALTER COLUMN exchange_key_id SET NOT NULL;
ALTER TABLE exchange_addresses DROP COLUMN IF EXISTS exchange_id;