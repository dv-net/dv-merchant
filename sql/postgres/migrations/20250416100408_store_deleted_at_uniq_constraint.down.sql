ALTER TABLE IF EXISTS stores
    ADD COLUMN IF NOT EXISTS slug varchar(255) default (concat('store_'::character varying,
                                                               (nextval('store_slug_seq'::regclass))::character varying))::character varying not null
        unique;
