DELETE FROM wallet_addresses a
USING wallet_addresses b
WHERE a.ctid < b.ctid
  AND a.currency_id = b.currency_id
  AND a.address = b.address
  AND a.dirty = false
  AND b.dirty = false;

CREATE UNIQUE INDEX unique_wallet_address
    ON wallet_addresses (currency_id, address)
    WHERE dirty = false;