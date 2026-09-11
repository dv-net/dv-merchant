CREATE INDEX CONCURRENTLY IF NOT EXISTS wallet_addresses_user_id_currency_id_idx
    ON wallet_addresses (user_id, currency_id);
