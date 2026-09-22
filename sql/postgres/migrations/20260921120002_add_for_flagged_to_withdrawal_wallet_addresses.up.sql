ALTER TABLE withdrawal_wallet_addresses
    ADD COLUMN for_flagged boolean NOT NULL DEFAULT false;
