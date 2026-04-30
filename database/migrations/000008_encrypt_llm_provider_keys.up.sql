-- Rename api_key to encrypted_key and change type to text for encrypted storage
ALTER TABLE llm_providers RENAME COLUMN api_key TO encrypted_key;
ALTER TABLE llm_providers ALTER COLUMN encrypted_key TYPE text;
