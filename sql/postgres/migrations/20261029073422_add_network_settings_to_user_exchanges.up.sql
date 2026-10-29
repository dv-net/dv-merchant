ALTER table user_exchanges
    ADD column proxy_url         TEXT        NULL,
    ADD column network_interface varchar(64) NULL;