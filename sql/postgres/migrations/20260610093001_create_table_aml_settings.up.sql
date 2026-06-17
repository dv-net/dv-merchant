CREATE TABLE store_aml_settings (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    store_id uuid NOT NULL UNIQUE REFERENCES stores(id),
    enabled boolean NOT NULL DEFAULT false,
    risk_threshold int NOT NULL DEFAULT 0,
    provider_slug varchar(255) DEFAULT null,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz DEFAULT now()
)