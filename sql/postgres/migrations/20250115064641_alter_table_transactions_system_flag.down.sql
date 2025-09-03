drop index if exists tx_is_system_idx;
alter table if exists transactions drop column if exists is_system;