ALTER TABLE api_keys DROP COLUMN IF EXISTS daily_limit;
ALTER TABLE api_keys DROP COLUMN IF EXISTS monthly_limit;
ALTER TABLE api_keys DROP COLUMN IF EXISTS daily_tokens;
ALTER TABLE api_keys DROP COLUMN IF EXISTS monthly_tokens;
