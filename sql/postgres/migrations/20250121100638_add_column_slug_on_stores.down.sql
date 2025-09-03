alter table stores
    drop column if exists slug;

alter table stores
    drop column if exists public_payment_form_enabled;

DROP SEQUENCE IF EXISTS store_slug_seq;