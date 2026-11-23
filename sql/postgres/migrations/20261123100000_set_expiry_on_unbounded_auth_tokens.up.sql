-- "remember me" logins used to be stored without expiry and lived forever.
-- Give every such user token the same 30-day ceiling that new logins now get.
UPDATE personal_access_tokens
SET expires_at = now() + interval '30 days',
    updated_at = now()
WHERE expires_at IS NULL
  AND tokenable_type = 'user'
  AND name = 'AuthToken';
