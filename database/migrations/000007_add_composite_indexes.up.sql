-- Add composite index for job cache lookup and performance
CREATE INDEX IF NOT EXISTS idx_jobs_content_hash_status ON jobs (content_hash, status);
CREATE INDEX IF NOT EXISTS idx_jobs_status_created ON jobs (status, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_job_step_logs_job_step ON job_step_logs (job_id, step_order);
