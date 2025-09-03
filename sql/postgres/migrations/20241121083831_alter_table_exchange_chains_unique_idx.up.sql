DROP INDEX IF EXISTS exchange_chains_unique_idx;
ALTER TABLE exchange_chains
    ADD CONSTRAINT exchange_chains_unique_idx
        UNIQUE (currency_id, slug);