CREATE TABLE IF NOT EXISTS invoices
(
    id                  uuid PRIMARY KEY        DEFAULT gen_random_uuid(),
    user_id             uuid           not null REFERENCES users (id) ON DELETE CASCADE,
    store_id            uuid           not null REFERENCES stores (id) ON DELETE CASCADE,
    order_id            varchar(32)    NOT NULL,
    expected_amount_usd numeric(28, 4) NOT NULL,
    received_amount_usd numeric(28, 4) NOT NULL DEFAULT 0,
    status              varchar(32)    NOT NULL DEFAULT 'pending',
    expires_at          timestamptz    NOT NULL,
    created_at          timestamptz    NOT NULL DEFAULT now(),
    updated_at          timestamptz    NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_invoices_store_order ON invoices (store_id, order_id);

CREATE INDEX IF NOT EXISTS idx_invoices_user_id ON invoices (user_id);
CREATE INDEX IF NOT EXISTS idx_invoices_store_id ON invoices (store_id);
CREATE INDEX IF NOT EXISTS idx_invoices_status ON invoices (status);
CREATE INDEX IF NOT EXISTS idx_invoices_expires_at ON invoices (expires_at);
CREATE INDEX IF NOT EXISTS idx_invoices_created_at ON invoices (created_at DESC);

CREATE TABLE IF NOT EXISTS invoice_addresses
(
    id                uuid PRIMARY KEY     DEFAULT gen_random_uuid(),
    invoice_id        uuid        NOT NULL REFERENCES invoices (id) ON DELETE CASCADE,
    wallet_address_id uuid        NOT NULL REFERENCES wallet_addresses (id) ON DELETE CASCADE,
    rate_at_creation  numeric(24, 8),
    created_at        timestamptz NOT NULL DEFAULT now(),
    updated_at        timestamptz NOT NULL DEFAULT now(),
    UNIQUE (invoice_id, wallet_address_id)
);

CREATE INDEX IF NOT EXISTS idx_invoice_addresses_wallet ON invoice_addresses (wallet_address_id);