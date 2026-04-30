-- Performance indexes for stats queries and dedup lookups
CREATE INDEX IF NOT EXISTS idx_job_step_logs_created_at ON job_step_logs(created_at);
CREATE INDEX IF NOT EXISTS idx_jobs_pipeline_created ON jobs(pipeline_id, created_at);
CREATE INDEX IF NOT EXISTS idx_entities_source_id ON entities(source_id);
CREATE INDEX IF NOT EXISTS idx_relations_from_to_type ON relations(from_entity_id, to_entity_id, type);
