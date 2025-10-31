create table if not exists wallet_address_status_history
(
    id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    wallet_address_id  uuid                           not null
        references wallet_addresses (id)
            on delete cascade,
    old_status varchar(32),
    new_status varchar(32)                    not null,
    changed_at timestamp        default now() not null
);