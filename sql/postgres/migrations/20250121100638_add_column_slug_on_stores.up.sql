CREATE SEQUENCE IF NOT EXISTS store_slug_seq START 1;

alter table stores
    add COLUMN if not exists slug varchar(255) not null default '';

alter table stores
    add COLUMN if not exists public_payment_form_enabled bool default false;

update stores
set slug = (lower(trim(regexp_replace(name, '\s+', '_', 'g'))) || nextval('store_slug_seq')::varchar);

alter table stores
    alter column slug set default (concat('store_'::varchar, nextval('store_slug_seq')::varchar)::varchar);

alter table stores add unique(slug);