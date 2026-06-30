-- jobs.api_key_id was defined on model.Job but never created by init schema
-- (dev sqlite+AutoMigrate masked it; prod migrations missed it).
-- Required for key-level job isolation (GetByUUID/List/cache filter by api_key_id).
ALTER TABLE jobs ADD COLUMN IF NOT EXISTS api_key_id BIGINT;
CREATE INDEX IF NOT EXISTS idx_jobs_api_key_id ON jobs(api_key_id);
