-- Add max_concurrency column to llm_providers for rate limiting
ALTER TABLE llm_providers ADD COLUMN max_concurrency INTEGER NOT NULL DEFAULT 10;
