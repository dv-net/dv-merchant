alter table currencies
    add is_stablecoin boolean not null default false;

update currencies set is_stablecoin = true where code = 'USDT' or code = 'USDC';