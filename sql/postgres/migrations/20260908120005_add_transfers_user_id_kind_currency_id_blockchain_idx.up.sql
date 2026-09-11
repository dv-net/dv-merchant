CREATE INDEX CONCURRENTLY IF NOT EXISTS transfers_user_id_kind_currency_id_blockchain_idx
    ON transfers (user_id, kind, currency_id, blockchain);
