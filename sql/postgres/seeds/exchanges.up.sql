insert into exchanges (id, slug, name, url, created_at, updated_at, is_active)
values
    ('0b08e649-b009-47c9-87c8-4dff132f5362', 'htx', 'HTX', 'https://api.huobi.pro', now(), now(), true),
    ('4975c77c-7591-4943-8f02-e6ad62172dab', 'okx', 'OKX', 'https://www.okx.com', now(), now(), true),
    ('9184d445-8eee-4020-916f-1c0c1cbff181', 'binance', 'BINANCE', 'https://api.binance.com', now(), now(), true),
    ('06fe1230-4689-4d33-909b-a9c15af64838', 'bitget', 'BITGET', 'https://api.bitget.com', now(), now(), true),
    ('2e07b084-ff69-4c09-bc9d-e93578260f67', 'kucoin', 'KUCOIN', 'https://api.kucoin.com', now(), now(), true),
    ('b868db0c-4d51-4842-ab6b-677f62c454fb', 'bybit', 'BYBIT', 'https://api.bybit.com', now(), now(), true),
    ('2495d8a2-435d-4031-8930-10e82aad9c64', 'gate', 'GATE', 'https://api.gateio.ws', now(), now(), true)
ON CONFLICT DO NOTHING;