ALTER TABLE user_aml_settings RENAME TO store_aml_settings;
ALTER TABLE store_aml_settings ADD COLUMN store_id uuid REFERENCES stores(id);
ALTER TABLE store_aml_settings DROP COLUMN user_id;
ALTER TABLE store_aml_settings ALTER COLUMN store_id SET NOT NULL;
ALTER TABLE store_aml_settings ADD CONSTRAINT store_aml_settings_store_id_key UNIQUE (store_id);