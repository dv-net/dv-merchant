ALTER TABLE aml_check_queue
    ADD COLUMN request_payload JSONB NOT NULL DEFAULT '{}';
