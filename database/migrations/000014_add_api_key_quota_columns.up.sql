-- Add quota columns to api_keys.
-- model.APIKey defines DailyLimit/MonthlyLimit/DailyTokens/MonthlyTokens and the
-- quota middleware reads them, but 000001_init_schema never created these columns
-- (dev-mode AutoMigrate masked the drift; migration-based prod deploy hit the gap).
ALTER TABLE api_keys ADD COLUMN IF NOT EXISTS daily_limit BIGINT NOT NULL DEFAULT 0;
ALTER TABLE api_keys ADD COLUMN IF NOT EXISTS monthly_limit BIGINT NOT NULL DEFAULT 0;
ALTER TABLE api_keys ADD COLUMN IF NOT EXISTS daily_tokens BIGINT NOT NULL DEFAULT 0;
ALTER TABLE api_keys ADD COLUMN IF NOT EXISTS monthly_tokens BIGINT NOT NULL DEFAULT 0;
