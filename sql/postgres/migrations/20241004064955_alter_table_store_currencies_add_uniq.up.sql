DELETE FROM store_currencies
WHERE (store_id, currency_id) IN (
    SELECT store_id, currency_id
    FROM (
             SELECT store_id, currency_id, ROW_NUMBER() OVER(PARTITION BY store_id, currency_id ORDER BY store_id) as rnum
             FROM store_currencies
         ) t
    WHERE t.rnum > 1
);
DROP INDEX IF EXISTS currency_store_currency_id_store_id_index;
ALTER TABLE store_currencies ADD CONSTRAINT uniq_currency_store_currency_id_store_id UNIQUE(store_id,currency_id);
