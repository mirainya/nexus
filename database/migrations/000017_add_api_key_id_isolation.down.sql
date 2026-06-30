DROP INDEX IF EXISTS idx_documents_api_key_id;
DROP INDEX IF EXISTS idx_entities_api_key_id;
DROP INDEX IF EXISTS idx_relations_api_key_id;
ALTER TABLE documents DROP COLUMN IF EXISTS api_key_id;
ALTER TABLE entities DROP COLUMN IF EXISTS api_key_id;
ALTER TABLE relations DROP COLUMN IF EXISTS api_key_id;
