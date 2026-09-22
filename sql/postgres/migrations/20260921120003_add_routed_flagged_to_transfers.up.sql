ALTER TABLE transfers
    ADD COLUMN routed_flagged boolean NOT NULL DEFAULT false;
