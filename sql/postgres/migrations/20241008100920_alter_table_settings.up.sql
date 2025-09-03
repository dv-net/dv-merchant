DROP INDEX IF EXISTS unique_name_when_model_is_null;
ALTER TABLE IF EXISTS settings ADD CONSTRAINT unique_name_model unique  NULLS NOT DISTINCT (name, model_id, model_type);