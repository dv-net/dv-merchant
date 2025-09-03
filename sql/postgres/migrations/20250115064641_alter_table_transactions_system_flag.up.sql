alter table transactions add column is_system boolean not null default false;
create index if not exists tx_is_system_idx on transactions(is_system);