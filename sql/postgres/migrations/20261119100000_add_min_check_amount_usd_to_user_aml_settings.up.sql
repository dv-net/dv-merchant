-- deposits with amount_usd below this value are not sent to the AML provider; 0 - check every deposit
ALTER TABLE user_aml_settings
    ADD COLUMN min_check_amount_usd numeric NOT NULL DEFAULT 0;
