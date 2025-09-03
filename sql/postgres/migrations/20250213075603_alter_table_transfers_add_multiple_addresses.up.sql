alter table if exists transfers
    add column if not exists from_addresses varchar(255)[] not null default ARRAY[]::varchar(255)[];
alter table if exists transfers
    add column if not exists to_addresses varchar(255)[] not null default ARRAY[]::varchar(255)[];

update transfers
set from_addresses = ARRAY [from_address],
    to_addresses   = ARRAY [to_address] where transfers.from_address != '';

ALTER TABLE IF EXISTS transfers
    DROP COLUMN IF EXISTS from_address;
ALTER TABLE IF EXISTS transfers
    DROP COLUMN IF EXISTS to_address;

