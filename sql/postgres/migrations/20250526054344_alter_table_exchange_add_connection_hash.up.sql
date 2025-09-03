ALTER TABLE IF EXISTS exchange_orders 
    ADD COLUMN IF NOT EXISTS exchange_connection_hash varchar(255);

ALTER TABLE IF EXISTS exchange_withdrawal_history 
    ADD COLUMN IF NOT EXISTS exchange_connection_hash varchar(255);