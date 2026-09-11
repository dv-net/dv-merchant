CREATE INDEX CONCURRENTLY IF NOT EXISTS transactions_to_address_currency_id_idx
    ON transactions (to_address, currency_id);
