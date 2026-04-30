ALTER TABLE llm_providers RENAME COLUMN encrypted_key TO api_key;
ALTER TABLE llm_providers ALTER COLUMN api_key TYPE varchar(500);
