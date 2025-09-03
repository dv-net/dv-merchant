CREATE TABLE IF NOT EXISTS user_exchanges (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    exchange_id uuid NOT NULL,
    user_id uuid NOT NULL,
    withdrawal_state VARCHAR(255) NOT NULL DEFAULT 'disabled',
    swap_state VARCHAR(255) NOT NULL DEFAULT 'disabled'
);

CREATE UNIQUE INDEX IF NOT EXISTS user_exchanges_unique_idx ON user_exchanges (exchange_id,user_id);

WITH active_user_exchanges AS (
    SELECT DISTINCT u.id AS user_id, e.id AS exchange_id
    FROM users u
             LEFT JOIN exchange_user_keys euk ON euk.user_id = u.id
             LEFT JOIN exchange_keys ek ON ek.id = euk.exchange_key_id
             LEFT JOIN exchanges e ON e.id = ek.exchange_id
    WHERE euk.exchange_key_id IS NOT NULL
)
INSERT INTO user_exchanges (user_id, exchange_id, withdrawal_state, swap_state)
SELECT user_id, exchange_id, 'disabled', 'disabled'
FROM active_user_exchanges
ON CONFLICT (user_id, exchange_id) DO NOTHING;