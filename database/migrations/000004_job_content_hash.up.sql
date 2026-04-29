ALTER TABLE jobs ADD COLUMN IF NOT EXISTS content_hash VARCHAR(64);
CREATE INDEX IF NOT EXISTS idx_jobs_content_hash ON jobs (content_hash) WHERE content_hash IS NOT NULL;
