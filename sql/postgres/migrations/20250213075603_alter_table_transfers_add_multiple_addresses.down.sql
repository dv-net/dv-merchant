DROP INDEX IF EXISTS uniq_addresses_to;
DROP INDEX IF EXISTS uniq_addresses_from;

ALTER TABLE IF EXISTS transfers
    ADD COLUMN IF NOT EXISTS from_address varchar(255);

ALTER TABLE IF EXISTS transfers
    ADD COLUMN IF NOT EXISTS to_address varchar(255);

UPDATE transfers
SET from_address = from_addresses[1],
    to_address = to_addresses[1]
WHERE array_length(from_addresses, 1) > 0 AND array_length(to_addresses, 1) > 0;

ALTER TABLE IF EXISTS transfers
    DROP COLUMN IF EXISTS from_addresses;

ALTER TABLE IF EXISTS transfers
    DROP COLUMN IF EXISTS to_addresses;
