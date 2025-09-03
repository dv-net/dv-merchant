create table if not exists logs
(
    id            uuid primary key            default gen_random_uuid(),
    log_type_slug varchar(100) not null,
    process_id    uuid         not null,
    level         varchar(100) not null,
    status        varchar(100) not null,
    message       text         not null,
    created_at    timestamp without time zone default current_timestamp
)