CREATE TABLE IF NOT EXISTS user_stores(
    id uuid default gen_random_uuid() primary key,
    user_id uuid not null,
    store_id uuid not null,
    created_at timestamp not null default now()
);

CREATE INDEX IF NOT EXISTS user_stores_index ON user_stores(user_id, store_id);

INSERT INTO user_stores (store_id, user_id) (SELECT s.id as store_id, s.user_id as user_id from stores s);
