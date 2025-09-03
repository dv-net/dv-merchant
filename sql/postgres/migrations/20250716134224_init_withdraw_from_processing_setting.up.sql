INSERT INTO settings (model_id, model_type, name, value, created_at, updated_at, is_mutable)
SELECT
    u.id AS model_id,
    'User' AS model_type,
    'withdraw_from_processing' AS name,
    'disabled' AS value,
    CURRENT_TIMESTAMP AS created_at,
    CURRENT_TIMESTAMP AS updated_at,
    true AS is_mutable
FROM users u
WHERE NOT EXISTS (
    SELECT 1
    FROM settings s
    WHERE s.model_id = u.id
      AND s.model_type = 'User'
      AND s.name = 'withdraw_from_processing'
)
ON CONFLICT ON CONSTRAINT unique_name_model
    DO NOTHING;