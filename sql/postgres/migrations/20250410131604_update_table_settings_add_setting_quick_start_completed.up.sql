INSERT INTO settings (model_id, model_type, name, value, created_at, is_mutable)
SELECT 
    id,
    'User',
    'quick_start_guide_status',
    'completed',
    NOW(),
    true
FROM users ON CONFLICT DO NOTHING;