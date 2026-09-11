CREATE INDEX CONCURRENTLY IF NOT EXISTS transactions_user_is_system_created_idx
    ON transactions (user_id, is_system, created_at_index DESC);
