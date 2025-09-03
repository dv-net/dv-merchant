DROP INDEX IF EXISTS tx_hash_tx_type_idx;

CREATE INDEX IF NOT EXISTS transfer_transactions_filter_idx ON transfer_transactions (created_at, tx_type, transfer_id);