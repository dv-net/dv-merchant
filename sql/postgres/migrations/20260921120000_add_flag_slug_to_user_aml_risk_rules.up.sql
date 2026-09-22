ALTER TABLE user_aml_risk_rules
    ADD COLUMN flag_slug varchar(50)
        CONSTRAINT user_aml_risk_rules_flag_slug_check
            CHECK (flag_slug IN ('sanctions', 'darknet_illicit', 'mixer_privacy', 'high_risk_exchange'))
        CONSTRAINT user_aml_risk_rules_flag_slug_action_check
            CHECK ((action = 'accept_and_flag') = (flag_slug IS NOT NULL));
