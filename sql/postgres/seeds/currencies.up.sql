insert into currencies (id, code, name, precision, is_fiat, blockchain, contract_address, withdrawal_min_balance, has_balance, created_at, updated_at, status, sort_order, min_confirmation, is_stablecoin, currency_label, token_label)
values
    ('BTC.Bitcoin', 'BTC', 'BTC', 8, false, 'bitcoin', 'btc', 0.00000000, true, 'now()', 'now()', true, 1, 1, false, 'Bitcoin', null),
    ('ETH.Ethereum', 'ETH', 'ETH', 6, false, 'ethereum', 'eth', 10.00000000, false, 'now()', 'now()',true, 1, 1, false, null, null),
    ('TRX.Tron', 'TRX', 'TRX', 6, false, 'tron', 'trx', 10.00000000, false, 'now()', 'now()', true, 1, 19, false, 'Tron', null),
    ('USD', 'USD', 'USD', 2, true, null, null, null, false, 'now()', 'now()', false, 1, null, false, null, null),
    ('USDT.Ethereum', 'USDT', 'USDT', 2, false, 'ethereum', '0xdac17f958d2ee523a2206206994597c13d831ec7', 10.00000000, false, 'now()', 'now()', true, 1, 1, true, 'Tether', 'ERC-20'),
    ('USDT.Tron', 'USDT', 'USDT', 2, false, 'tron', 'TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t', 0.50000000, true, 'now()', 'now()', true, 1, 19, true, 'Tether', 'TRC-20'),
    ('USDC.Ethereum', 'USDC', 'USDC', 2, false, 'ethereum', '0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48', 1.00000000, true, 'now()', 'now()', true, 1, 1, true, 'Circle USD', 'ERC-20'),
    ('DAI.Ethereum', 'DAI', 'DAI', 2, false, 'ethereum', '0x6b175474e89094c44da98b954eedeac495271d0f', 1.00000000, true, 'now()', 'now()', true, 1, 1, true, null, 'ERC-20'),
    ('LTC.Litecoin', 'LTC', 'LTC', 8, false, 'litecoin', 'ltc', 0.00000000, true, 'now()', 'now()', true, 1, 1, false, 'Litecoin', null),
    ('BCH.Bitcoincash', 'BCH', 'BCH', 8, false, 'bitcoincash', 'bch', 0.00000000, true, 'now()', 'now()', true, 1, 1, false, 'Bitcoincash', null),
    ('BNB.BNBSmartChain', 'BNB', 'BNB', 8, false, 'bsc', 'bnb', 0.00000000, true, 'now()', 'now()', true, 1, 1, false, 'BSC', null),
    ('USDT.BNBSmartChain', 'USDT', 'USDT', 8, false, 'bsc', '0x55d398326f99059ff775485246999027b3197955', 0.00000000, true, 'now()', 'now()', true, 1, 1, true, 'Tether', 'BEP-20'),
    ('USDC.BNBSmartChain', 'USDC', 'USDC', 8, false, 'bsc', '0x8ac76a51cc950d9822d68b83fe1ad97b32cd580d', 0.00000000, true, 'now()', 'now()', true, 1, 1, true, 'Circle USD', 'BEP-20'),
    ('DAI.BNBSmartChain', 'DAI', 'DAI', 8, false, 'bsc', '0x1af3f329e8be154074d8769d1ffa4ee058b1dbc3', 0.00000000, true, 'now()', 'now()', true, 1, 1, true, null, 'BEP-20'),
    ('POL.Polygon', 'POL', 'POL', 8, false, 'polygon', 'pol', 0.00000000, true, 'now()', 'now()', true, 1, 1, false, 'Polygon', null),
    ('USDT.Polygon', 'USDT', 'USDT', 8, false, 'polygon', '0xc2132d05d31c914a87c6611c10748aeb04b58e8f', 0.00000000, true, 'now()', 'now()', true, 1, 1, true, 'Tether', null),
    ('USDC.Polygon', 'USDC', 'USDC', 8, false, 'polygon', '0x3c499c542cef5e3811e1192ce70d8cc03d5c3359', 0.00000000, true, 'now()', 'now()', true, 1, 1, true, 'Circle USD', null),
    ('DAI.Polygon', 'DAI', 'DAI', 8, false, 'polygon', '0x8f3cf7ad23cd3cadbd9735aff958023239c6a063', 0.00000000, true, 'now()', 'now()', true, 1, 1, true, null, null),
    ('DOGE.Dogecoin', 'DOGE', 'DOGE', 8, false, 'dogecoin', 'doge', 0.00000000, true, 'now()', 'now()', true, 1, 1, false, 'Dogecoin', null),
    ('ETH.Arbitrum', 'ETH', 'ETH', 8, false, 'arbitrum', 'eth', 0.00000000, true, 'now()', 'now()', true, 1, 19, false, null, null),
    ('USDT.Arbitrum', 'USDT', 'USDT', 2, false, 'arbitrum', '0xfd086bc7cd5c481dcc9c85ebe478a1c0b69fcbb9', 0.00000000, true, 'now()', 'now()', true, 1, 19, true, 'Tether', null),
    ('USDC.Arbitrum', 'USDC', 'USDC', 2, false, 'arbitrum', '0xaf88d065e77c8cc2239327c5edb3a432268e5831', 0.00000000, true, 'now()', 'now()', true, 1, 19, true, 'Circle USD', null),
    ('DAI.Arbitrum', 'DAI', 'DAI', 2, false, 'arbitrum', '0xda10009cbd5d07dd0cecc66161fc93d7c9000da1', 0.00000000, true, 'now()', 'now()', true, 1, 19, true, null, null)
ON CONFLICT (id) DO UPDATE SET
    currency_label = EXCLUDED.currency_label,
    token_label = EXCLUDED.token_label;