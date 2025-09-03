ALTER TABLE withdrawal_from_processing_wallets
    ADD COLUMN IF NOT EXISTS request_id VARCHAR(64) UNIQUE;