CREATE TABLE user_aml_risk_rules
(
    id            uuid PRIMARY KEY      DEFAULT gen_random_uuid(),
    user_id       uuid         NOT NULL REFERENCES users (id),
    provider_slug varchar(255) NOT NULL,
    risk_type     varchar(255) NOT NULL, -- 'TOTAL_RISK_SCORE' | 'SUM_OF_SIGNALS' 
    enabled       boolean      NOT NULL DEFAULT false,
    threshold     numeric      NOT NULL,
    action        varchar(50)  NOT NULL, -- 'reject' | 'accept_and_flag'
    created_at    timestamptz  NOT NULL DEFAULT now(),
    updated_at    timestamptz           DEFAULT now(),
    UNIQUE (user_id, provider_slug, risk_type)
);

INSERT INTO user_aml_risk_rules (user_id, risk_type, enabled, threshold, action, provider_slug)
SELECT user_id, 'TOTAL_RISK_SCORE', enabled, risk_threshold, 'reject', provider_slug
FROM user_aml_settings;

ALTER TABLE user_aml_settings DROP COLUMN risk_threshold;