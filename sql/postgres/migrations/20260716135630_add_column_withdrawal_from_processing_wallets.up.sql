ALTER TABLE withdrawal_from_processing_wallets
    ADD COLUMN blocked_by_processing_error boolean NOT NULL DEFAULT false;