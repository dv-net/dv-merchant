ALTER TABLE user_aml_settings ADD COLUMN enabled boolean NOT NULL DEFAULT false;
ALTER TABLE user_aml_settings ADD COLUMN risk_threshold int NOT NULL DEFAULT 0;

UPDATE user_aml_settings uas
SET enabled = r.enabled,
    risk_threshold = r.threshold::int
FROM user_aml_risk_rules r
WHERE r.user_id = uas.user_id AND r.risk_type = 'TOTAL_RISK_SCORE';

DROP TABLE user_aml_risk_rules;