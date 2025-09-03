BEGIN;

INSERT INTO settings (id, name, value, created_at, updated_at, is_mutable)
SELECT
    gen_random_uuid() AS id,
    'withdraw_from_processing' AS name,
    (
        SELECT value
        FROM settings
        WHERE name = 'withdraw_from_processing'
          AND model_type = 'User'
        GROUP BY value
        ORDER BY COUNT(*) DESC
        LIMIT 1
    ) AS value,
    CURRENT_TIMESTAMP AS created_at,
    CURRENT_TIMESTAMP AS updated_at,
    true AS is_mutable
ON CONFLICT ON CONSTRAINT unique_name_model
    DO NOTHING;

DELETE FROM settings
WHERE name = 'withdraw_from_processing'
  AND model_type = 'User';

COMMIT;