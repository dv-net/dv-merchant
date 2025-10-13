CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_wallet_addresses_lookup
    ON wallet_addresses (wallet_id, currency_id, created_at DESC)
    WHERE deleted_at IS NULL;
