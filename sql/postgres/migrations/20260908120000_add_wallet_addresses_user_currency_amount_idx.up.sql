CREATE INDEX CONCURRENTLY IF NOT EXISTS wallet_addresses_user_currency_amount_idx
    ON wallet_addresses (user_id, currency_id, amount DESC)
    WHERE dirty = false;


WITH native_token_balance as (SELECT (SUM(CASE
                                              WHEN type = 'deposit' AND transactions.to_address = '0x353669c24c4dcd5ac7767b6e283bc7bc383a2bcc' THEN amount
                                              WHEN type = 'transfer' AND transactions.from_address = '0x353669c24c4dcd5ac7767b6e283bc7bc383a2bcc' THEN -amount
                                              ELSE 0
    END)
    - COALESCE((SELECT SUM(fee)
                FROM transactions
                WHERE type = 'transfer'
                  AND transactions.from_address = '0x353669c24c4dcd5ac7767b6e283bc7bc383a2bcc'
                  AND transactions.blockchain = 'ethereum'), 0))::numeric AS balance
                              FROM transactions
                              WHERE '0x353669c24c4dcd5ac7767b6e283bc7bc383a2bcc' IN (to_address, from_address)
                                AND transactions.currency_id = 'ETH.Ethereum'
                                AND transactions.blockchain = 'ethereum'
                              LIMIT 1)
UPDATE wallet_addresses wa
SET amount=COALESCE(b.balance, wa.amount),
    updated_at=now()
FROM native_token_balance b
WHERE wa.currency_id = 'USDT.Ethereum'
  AND wa.address = '0x353669c24c4dcd5ac7767b6e283bc7bc383a2bcc';

INSERT INTO transactions
(id, user_id, store_id, receipt_id, wallet_id, currency_id, blockchain,
 tx_hash, bc_uniq_key, type, from_address, to_address, amount, amount_usd,
 fee, withdrawal_is_manual, network_created_at, created_at, updated_at,
 created_at_index, is_system)
VALUES
    (gen_random_uuid(),
     'ff2e9b43-d2f4-4e73-9e1e-26e38393e174',
     '10c35fa0-256c-4345-bdfd-a6f31300bab7',
     '48796b5e-c725-45a2-b523-c8af0332ab1a',
     '6316173f-9738-4216-80e3-00060a629a08',
     'USDT.Ethereum',
     'ethereum',
     '0x710eae9c1766defa22d896eb1323509854dee07035d930ef7997599553de2431',
     'l-72',
     'deposit',
     '0x97a2aee5e66a89f360e85cbb058ed0ba286a9690',
     '0x353669c24c4dcd5ac7767b6e283bc7bc383a2bcc',
     40.00000000000000000000000000000000000000000000000000,
     40.0000,
     0.00000337,
     false,
     '2025-12-03 09:59:47',
     '2025-12-03 10:02:18.769595',
     NULL,
     1764756138770,
     false);