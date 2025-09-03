ALTER TABLE IF EXISTS exchange_orders 
	DROP COLUMN IF EXISTS exchange_connection_hash;

ALTER TABLE IF EXISTS exchange_withdrawal_history
	DROP COLUMN IF EXISTS exchange_connection_hash;