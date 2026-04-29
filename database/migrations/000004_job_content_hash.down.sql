DROP INDEX IF EXISTS idx_jobs_content_hash;
ALTER TABLE jobs DROP COLUMN IF EXISTS content_hash;
