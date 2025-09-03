ALTER TABLE currencies 
DROP COLUMN IF EXISTS currency_label,
DROP COLUMN IF EXISTS token_label;