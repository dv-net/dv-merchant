ALTER TABLE store_aml_settings RENAME TO user_aml_settings;
ALTER TABLE user_aml_settings ADD COLUMN user_id uuid REFERENCES users(id);

UPDATE user_aml_settings uas
SET user_id = s.user_id
    FROM stores s
WHERE s.id = uas.store_id;

ALTER TABLE user_aml_settings ALTER COLUMN store_id DROP NOT NULL;

WITH consolidated AS (
    SELECT
        user_id,
        BOOL_OR(enabled) AS enabled,
        MIN(risk_threshold) AS risk_threshold,
        (ARRAY_AGG(provider_slug ORDER BY risk_threshold ASC, enabled DESC, updated_at DESC NULLS LAST))[1] AS provider_slug,
    MIN(created_at) AS created_at
FROM user_aml_settings
GROUP BY user_id
    )
INSERT INTO user_aml_settings (id, user_id, enabled, risk_threshold, provider_slug, created_at, updated_at)
SELECT gen_random_uuid(), user_id, enabled, risk_threshold, provider_slug, created_at, now()
FROM consolidated;

DELETE FROM user_aml_settings WHERE store_id IS NOT NULL;

ALTER TABLE user_aml_settings DROP COLUMN store_id;
ALTER TABLE user_aml_settings ALTER COLUMN user_id SET NOT NULL;
ALTER TABLE user_aml_settings ADD CONSTRAINT user_aml_settings_user_id_key UNIQUE (user_id);