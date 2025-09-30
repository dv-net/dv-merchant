ALTER TABLE currencies 
    ADD COLUMN is_new_store_default BOOLEAN NOT NULL DEFAULT FALSE;