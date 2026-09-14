CREATE INDEX CONCURRENTLY IF NOT EXISTS wallet_addresses_user_currency_amount_idx
    ON wallet_addresses (user_id, currency_id, amount DESC)
    WHERE dirty = false;