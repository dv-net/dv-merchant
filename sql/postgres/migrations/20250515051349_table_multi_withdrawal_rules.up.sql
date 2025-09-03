CREATE TABLE IF NOT EXISTS multi_withdrawal_rules(
                                                     id uuid primary key NOT NULL DEFAULT gen_random_uuid(),
                                                     withdrawal_wallet_id uuid NOT NULL UNIQUE,
                                                     mode VARCHAR(255) NOT NULL DEFAULT 'random_wallet',
                                                     manual_address VARCHAR(255) default null,
                                                     created_at TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP,
                                                     updated_at TIMESTAMP WITHOUT TIME ZONE DEFAULT NULL
);