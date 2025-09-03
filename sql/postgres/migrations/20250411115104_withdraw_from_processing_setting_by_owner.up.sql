BEGIN;

INSERT INTO settings (model_id, model_type, name, value, created_at, updated_at, is_mutable)
SELECT
    u.id AS model_id,
    'User' AS model_type,
    'withdraw_from_processing' AS name,
    COALESCE(
            (SELECT s.value
             FROM settings s
             WHERE s.name = 'withdraw_from_processing'
               AND s.model_id IS NULL
               AND s.model_type IS NULL),
            'disabled'
    ) AS value,
    CURRENT_TIMESTAMP AS created_at,
    CURRENT_TIMESTAMP AS updated_at,
    true AS is_mutable
FROM users u
ON CONFLICT ON CONSTRAINT unique_name_model
    DO NOTHING;

DELETE FROM settings
WHERE name = 'withdraw_from_processing'
  AND model_id IS NULL
  AND model_type IS NULL;

COMMIT;