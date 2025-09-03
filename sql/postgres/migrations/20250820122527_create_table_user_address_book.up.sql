CREATE TABLE IF NOT EXISTS user_address_book(
    id uuid primary key NOT NULL DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL,
    address VARCHAR(255) NOT NULL,
    currency_id VARCHAR(255) NULL,
    name VARCHAR(255) DEFAULT NULL,
    tag VARCHAR(255) DEFAULT NULL,
    blockchain VARCHAR(255) NULL,
    type VARCHAR(255) NOT NULL DEFAULT 'simple',
    submitted_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NULL,
    deleted_at TIMESTAMP WITH TIME ZONE DEFAULT NULL,
    UNIQUE (user_id, address, currency_id, type)
);

CREATE INDEX IF NOT EXISTS idx_user_address_book_user_id ON user_address_book(user_id);
CREATE INDEX IF NOT EXISTS idx_user_address_book_blockchain ON user_address_book(blockchain);
CREATE INDEX IF NOT EXISTS idx_user_address_book_address ON user_address_book(address);
CREATE INDEX IF NOT EXISTS idx_user_address_book_currency_id ON user_address_book(currency_id);