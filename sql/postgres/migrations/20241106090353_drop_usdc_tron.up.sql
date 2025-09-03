BEGIN TRANSACTION;

DELETE
FROM transactions
WHERE currency_id = 'USDC.Tron';
DELETE

FROM withdrawal_from_processing_wallets
WHERE currency_id = 'USDC.Tron';

DELETE
FROM transfers
WHERE currency_id = 'USDC.Tron';

DELETE
FROM unconfirmed_transactions
WHERE currency_id = 'USDC.Tron';

DELETE
FROM store_currencies
WHERE currency_id = 'USDC.Tron';

DELETE
FROM receipts
WHERE currency_id = 'USDC.Tron';

DELETE
FROM withdrawal_wallet_addresses
WHERE withdrawal_wallet_addresses.withdrawal_wallet_id in (SELECT id FROM withdrawal_wallets WHERE currency_id = 'USDC.Tron');

DELETE
FROM wallet_addresses
WHERE currency_id = 'USDC.Tron';

DELETE
FROM withdrawal_wallets
WHERE currency_id = 'USDC.Tron';

UPDATE stores SET currency_id = 'USDT.Tron' WHERE currency_id = 'USDC.Tron';

DELETE FROM currencies WHERE id = 'USDC.Tron';

COMMIT;