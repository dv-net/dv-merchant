ALTER TABLE store_aml_settings
    ADD COLUMN ignored_signal_categories jsonb NOT NULL DEFAULT '[]'::jsonb;