DROP INDEX IF EXISTS idx_jobs_api_key_id;
ALTER TABLE jobs DROP COLUMN IF EXISTS api_key_id;
