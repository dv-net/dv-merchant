CREATE INDEX CONCURRENTLY IF NOT EXISTS wallet_addresses_user_currency_amount_idx
    ON wallet_addresses (user_id, currency_id, amount DESC)
    WHERE dirty = false;

CREATE INDEX CONCURRENTLY IF NOT EXISTS wallet_addresses_user_id_currency_id_idx
    ON wallet_addresses (user_id, currency_id);

CREATE INDEX CONCURRENTLY IF NOT EXISTS transactions_user_is_system_created_idx
    ON transactions (user_id, is_system, created_at_index DESC);

CREATE INDEX CONCURRENTLY IF NOT EXISTS aml_checks_transaction_id_idx
    ON aml_checks (transaction_id);

CREATE INDEX CONCURRENTLY IF NOT EXISTS transfers_user_id_kind_currency_id_blockchain_idx
    ON transfers (user_id, kind, currency_id, blockchain);

CREATE INDEX CONCURRENTLY IF NOT EXISTS transactions_to_address_currency_id_idx
    ON transactions (to_address, currency_id);

CREATE INDEX CONCURRENTLY IF NOT EXISTS transactions_from_address_currency_id_blockchain_idx
    ON transactions (from_address, currency_id, blockchain);
