ALTER TABLE blocked_transactions
    ADD COLUMN block_reasons jsonb NOT NULL DEFAULT '[]'::jsonb;
