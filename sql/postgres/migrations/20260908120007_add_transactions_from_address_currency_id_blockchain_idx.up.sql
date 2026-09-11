CREATE INDEX CONCURRENTLY IF NOT EXISTS transactions_from_address_currency_id_blockchain_idx
    ON transactions (from_address, currency_id, blockchain);
