ALTER TABLE wallet_addresses
    ADD COLUMN risk_flags jsonb NOT NULL DEFAULT '[]'::jsonb;
